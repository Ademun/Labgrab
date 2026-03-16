package usecase

import (
	"context"
	"errors"
	"fmt"
	"labgrab/internal/auth"
	"labgrab/internal/booking"
	"labgrab/internal/event"
	"labgrab/internal/shared/domain"
	"labgrab/internal/subscription"
	"labgrab/internal/telegram"
	"labgrab/internal/user"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// enrollmentTotal counts enrollment attempts by outcome.
	// outcome labels: "success", "auth_nil_creds", "enroll_error",
	//                 "fresh_schedule_empty", "subscription_inactive",
	//                 "pick_slot_failed", "missing_user_fields"
	enrollmentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "labgrab",
		Subsystem: "event_usecase",
		Name:      "enrollments_total",
		Help:      "Total enrollment attempts broken down by outcome.",
	}, []string{"outcome"})

	// cycleDurationSeconds measures the wall-clock time of a full Exec() call.
	// This is the critical metric for detecting overlapping runs.
	cycleDurationSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "labgrab",
		Subsystem: "event_usecase",
		Name:      "cycle_duration_seconds",
		Help:      "Wall-clock duration of a full ProcessEvents Exec() call.",
		Buckets:   []float64{5, 15, 30, 60, 120, 300, 600, 900, 1200},
	})

	// enrollChFill tracks how full the enrollCh buffer is at the moment of send.
	// Consistently high values mean event workers outpace user workers.
	enrollChFill = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "labgrab",
		Subsystem: "event_usecase",
		Name:      "enroll_channel_fill",
		Help:      "Current number of tasks buffered in enrollCh (cap=50).",
	})

	// activeUserWorkers is a gauge that tracks concurrent userWorker goroutines.
	// Use this to justify / tune the global semaphore size.
	activeUserWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "labgrab",
		Subsystem: "event_usecase",
		Name:      "active_user_workers",
		Help:      "Number of userWorker goroutines currently running.",
	})

	// mcmfConflictsTotal counts events that triggered MCMF conflict resolution.
	// High values = many users share the same subscription patterns.
	mcmfConflictsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "labgrab",
		Subsystem: "event_usecase",
		Name:      "mcmf_conflicts_total",
		Help:      "Number of events that required MCMF conflict resolution.",
	})

	// subscriptionInactiveTotal counts how often the re-check in userWorker
	// catches a subscription that became inactive between eventWorker and userWorker.
	// This is the double-enrollment guard metric.
	subscriptionInactiveTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "labgrab",
		Subsystem: "event_usecase",
		Name:      "subscription_inactive_recheck_total",
		Help:      "Times a subscription was inactive at userWorker re-check (double-enroll guard).",
	})
)

type ProcessEventsUsecase struct {
	EventSvc        *event.Service
	UserSvc         *user.Service
	AuthSvc         *auth.Service
	BookingSvc      *booking.Service
	SubscriptionSvc *subscription.Service
	TelegramSvc     *telegram.Service
}

type enrollTask struct {
	eventRes *event.GetEventsRes
	sub      subscription.GetMatchingSubscriptionsRes
	userInfo *user.GetUserRes
}

func (uc *ProcessEventsUsecase) Exec(ctx context.Context) error {
	start := time.Now()
	defer func() {
		cycleDurationSeconds.Observe(time.Since(start).Seconds())
	}()

	var clientCookies string
	events, err := uc.EventSvc.GetCurrentEvents(ctx, &clientCookies)
	if err != nil {
		return fmt.Errorf("event usecase: exec: get current events: %w", err)
	}

	enrollCh := make(chan enrollTask, 50)
	errCh := make(chan error)

	var eventWg sync.WaitGroup
	for range 3 {
		eventWg.Add(1)
		go func() {
			defer eventWg.Done()
			uc.eventWorker(ctx, events, enrollCh, errCh)
		}()
	}

	go func() {
		eventWg.Wait()
		close(enrollCh)
	}()

	go func() {
		userChannels := make(map[uuid.UUID]chan enrollTask)
		var userWg sync.WaitGroup
		for task := range enrollCh {
			// Track live buffer fill so we can detect backpressure.
			enrollChFill.Set(float64(len(enrollCh)))

			userID := task.sub.UserUUID
			if _, ok := userChannels[userID]; !ok {
				ch := make(chan enrollTask, 10)
				userChannels[userID] = ch
				userWg.Add(1)
				go func(userCh <-chan enrollTask) {
					defer userWg.Done()
					activeUserWorkers.Inc()
					uc.userWorker(ctx, userCh, errCh)
					activeUserWorkers.Dec()
				}(ch)
			}
			userChannels[userID] <- task
		}
		for _, ch := range userChannels {
			close(ch)
		}
		userWg.Wait()
		close(errCh)
	}()

	var collected error
	for err := range errCh {
		collected = errors.Join(collected, err)
	}

	if collected != nil {
		return fmt.Errorf("event usecase: exec: %w", collected)
	}

	return nil
}

func (uc *ProcessEventsUsecase) eventWorker(
	ctx context.Context,
	events <-chan *event.GetEventsRes,
	enrollCh chan<- enrollTask,
	errCh chan<- error,
) {
	for e := range events {
		if e.Err != nil {
			errCh <- fmt.Errorf("event usecase: event worker: event error: %w", e.Err)
			continue
		}

		subs, err := uc.SubscriptionSvc.GetMatchingSubscriptions(ctx, &subscription.GetMatchingSubscriptionsReq{
			Type:       e.Data.Type,
			Topic:      e.Data.Topic,
			Number:     e.Data.Number,
			Auditorium: e.Data.Auditorium,
			Schedule:   e.Data.Schedule,
		})
		if err != nil {
			errCh <- fmt.Errorf("event usecase: event worker: get matching subscriptions: %w", err)
			continue
		}

		if len(subs) == 0 {
			continue
		}

		subUsers := make([]*user.GetUserRes, len(subs))
		for i := range subs {
			userInfo, err := uc.UserSvc.GetUser(ctx, subs[i].UserUUID)
			if err != nil {
				errCh <- fmt.Errorf("event usecase: event worker: get user: %w", err)
				continue
			}
			subUsers[i] = userInfo
		}

		enrollIdxs := make([]int, 0, len(subs))
		for i := range subs {
			if subUsers[i] == nil {
				continue
			}

			if !subs[i].AutoEnroll {
				err := uc.TelegramSvc.NotifyEvent(ctx, telegram.NotifyEventReq{
					UserID:        subUsers[i].TelegramID,
					LabName:       e.Data.Name,
					LabType:       e.Data.Type,
					LabTopic:      e.Data.Topic,
					LabNumber:     e.Data.Number,
					LabAuditorium: e.Data.Auditorium,
					Spot:          e.Data.Spot,
					Schedule:      subs[i].Schedule,
					PageURL:       e.Data.Link,
				})
				if err != nil {
					errCh <- fmt.Errorf("event usecase: event worker: notify event: %w", err)
				}
				continue
			}

			filteredSchedule, err := uc.BookingSvc.FilterSchedule(ctx, &booking.FilterScheduleReq{
				UserUUID: subs[i].UserUUID,
				Schedule: subs[i].Schedule,
				Type:     e.Data.Type,
				Topic:    e.Data.Topic,
				Number:   e.Data.Number,
			})
			if err != nil {
				errCh <- fmt.Errorf("event usecase: event worker: filter schedule: %w", err)
				continue
			}

			subs[i].Schedule = filteredSchedule
			enrollIdxs = append(enrollIdxs, i)
		}

		if len(enrollIdxs) == 0 {
			continue
		}

		keyMap := make(map[time.Time]int)
		conflictingDates := make([]time.Time, 0)
		for _, i := range enrollIdxs {
			for d := range subs[i].Schedule {
				utcD := d.UTC()
				keyMap[utcD]++
				if keyMap[utcD] == 2 {
					conflictingDates = append(conflictingDates, d)
				}
			}
		}

		// Count events that actually needed conflict resolution.
		if len(conflictingDates) > 0 {
			mcmfConflictsTotal.Inc()
		}

		subIndex := make(map[uuid.UUID]int, len(enrollIdxs))
		for _, i := range enrollIdxs {
			subIndex[subs[i].SubscriptionUUID] = i
		}

		for _, date := range conflictingDates {
			constraints := make(map[uuid.UUID][]domain.Lesson)
			for _, i := range enrollIdxs {
				if l, ok := subs[i].Schedule[date]; ok {
					lessons := make([]domain.Lesson, 0, len(l))
					for k := range l {
						lessons = append(lessons, k)
					}
					constraints[subs[i].SubscriptionUUID] = lessons
				}
			}

			resolved := uc.SubscriptionSvc.ResolveConflicts(constraints)

			for id, lessons := range resolved {
				if i, ok := subIndex[id]; ok {
					subs[i].Schedule[date] = make(map[domain.Lesson][]string)
					for _, lesson := range lessons {
						subs[i].Schedule[date][lesson] = make([]string, 0)
					}
				}
			}
		}

		for _, i := range enrollIdxs {
			if len(subs[i].Schedule) == 0 {
				continue
			}
			enrollCh <- enrollTask{
				eventRes: e,
				sub:      subs[i],
				userInfo: subUsers[i],
			}
		}
	}
}

func (uc *ProcessEventsUsecase) userWorker(
	ctx context.Context,
	tasks <-chan enrollTask,
	errCh chan<- error,
) {
	for task := range tasks {
		if task.userInfo.Name == nil || task.userInfo.Surname == nil || task.userInfo.Patronymic == nil || task.userInfo.GroupCode == nil {
			enrollmentTotal.WithLabelValues("missing_user_fields").Inc()
			continue
		}

		sub, err := uc.SubscriptionSvc.GetSubscription(ctx, task.sub.SubscriptionUUID)
		if err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: get subscription: %w", err)
			continue
		}
		if sub.Status != subscription.StatusActive {
			// Double-enrollment guard fired — subscription closed between
			// eventWorker dispatch and userWorker execution.
			subscriptionInactiveTotal.Inc()
			enrollmentTotal.WithLabelValues("subscription_inactive").Inc()
			continue
		}

		creds, err := uc.AuthSvc.GetUserCredentials(ctx, task.sub.UserUUID)
		if err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: get user credentials: %w", err)
			continue
		}
		if creds.Session == nil || creds.Cookies == nil {
			enrollmentTotal.WithLabelValues("auth_nil_creds").Inc()
			errCh <- fmt.Errorf("event usecase: user worker: get user credentials: nil session or cookies")
			continue
		}

		err = uc.BookingSvc.LoadClientBookings(ctx, &booking.LoadClientBookingsReq{
			UserUUID: task.sub.UserUUID,
			Session:  *creds.Session,
			Cookies:  *creds.Cookies,
		})
		if err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: load client booking: %w", err)
			continue
		}

		freshSchedule, err := uc.BookingSvc.FilterSchedule(ctx, &booking.FilterScheduleReq{
			UserUUID: task.sub.UserUUID,
			Schedule: task.sub.Schedule,
			Type:     task.eventRes.Data.Type,
			Topic:    task.eventRes.Data.Topic,
			Number:   task.eventRes.Data.Number,
		})
		if err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: filter schedule: %w", err)
			continue
		}
		if len(freshSchedule) == 0 {
			// Slot was taken between eventWorker dispatch and now.
			enrollmentTotal.WithLabelValues("fresh_schedule_empty").Inc()
			continue
		}

		selectedDate, selectedLesson, ok := pickRandomSlot(freshSchedule)
		if !ok {
			enrollmentTotal.WithLabelValues("pick_slot_failed").Inc()
			continue
		}

		lessonTime := domain.LessonLookup[int(selectedLesson)]
		targetTime := time.Date(selectedDate.Year(), selectedDate.Month(), selectedDate.Day(), lessonTime.Start.Hour(), lessonTime.Start.Minute(), 0, 0, time.UTC)

		_, err = uc.EventSvc.Enroll(ctx, &event.EnrollmentReq{
			UserUUID:    task.sub.UserUUID,
			EventID:     task.eventRes.Data.ID,
			ServiceID:   task.eventRes.Data.ServiceID,
			Time:        targetTime,
			Name:        *task.userInfo.Name,
			Surname:     *task.userInfo.Surname,
			Patronymic:  *task.userInfo.Patronymic,
			Group:       *task.userInfo.GroupCode,
			PhoneNumber: creds.DikidiPhoneNumber,
			Session:     *creds.Session,
			Cookies:     *creds.Cookies,
		})
		if err != nil {
			enrollmentTotal.WithLabelValues("enroll_error").Inc()
			errCh <- fmt.Errorf("event usecase: user worker: enroll: %w", err)
			continue
		}

		enrollmentTotal.WithLabelValues("success").Inc()

		if err = uc.SubscriptionSvc.CloseSubscription(ctx, task.sub.SubscriptionUUID); err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: update subscription: %w", err)
		}

		if err = uc.TelegramSvc.NotifyEnrollment(ctx, telegram.NotifyEnrollmentReq{
			UserID:        task.userInfo.TelegramID,
			LabName:       task.eventRes.Data.Name,
			LabType:       task.eventRes.Data.Type,
			LabTopic:      task.eventRes.Data.Topic,
			LabNumber:     task.eventRes.Data.Number,
			LabAuditorium: task.eventRes.Data.Auditorium,
			Spot:          task.eventRes.Data.Spot,
			Date:          selectedDate,
			Lesson:        selectedLesson,
		}); err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: notify enrollment: %w", err)
		}
	}
}

func pickRandomSlot(schedule map[time.Time]map[domain.Lesson][]string) (time.Time, domain.Lesson, bool) {
	validDates := make([]time.Time, 0, len(schedule))
	for d, lessons := range schedule {
		if len(lessons) > 0 {
			validDates = append(validDates, d)
		}
	}
	if len(validDates) == 0 {
		return time.Time{}, 0, false
	}

	selectedDate := validDates[rand.IntN(len(validDates))]

	lessonKeys := make([]domain.Lesson, 0, len(schedule[selectedDate]))
	for k := range schedule[selectedDate] {
		lessonKeys = append(lessonKeys, k)
	}
	if len(lessonKeys) == 0 {
		return time.Time{}, 0, false
	}

	return selectedDate, lessonKeys[rand.IntN(len(lessonKeys))], true
}
