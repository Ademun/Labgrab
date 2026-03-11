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
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/google/uuid"
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
	var clientCookies string
	events, err := uc.EventSvc.GetCurrentEvents(ctx, &clientCookies)
	if err != nil {
		return fmt.Errorf("event usecase: exec: get current events: %w", err)
	}

	enrollCh := make(chan enrollTask, 100)
	errCh := make(chan error, 100)

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

	userChannels := make(map[uuid.UUID]chan enrollTask)
	var userWg sync.WaitGroup

	for task := range enrollCh {
		userID := task.sub.UserUUID
		if _, ok := userChannels[userID]; !ok {
			ch := make(chan enrollTask, 50)
			userChannels[userID] = ch
			userWg.Add(1)
			go func(userCh <-chan enrollTask) {
				defer userWg.Done()
				uc.userWorker(ctx, userCh, errCh)
			}(ch)
		}
		userChannels[userID] <- task
	}

	for _, ch := range userChannels {
		close(ch)
	}

	userWg.Wait()
	close(errCh)

	var collected error
	for err := range errCh {
		fmt.Println(err)
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

		subIndex := make(map[uuid.UUID]int, len(enrollIdxs))
		for _, i := range enrollIdxs {
			subIndex[subs[i].SubscriptionUUID] = i
		}

		for _, date := range conflictingDates {
			leftSide := make(map[uuid.UUID][]domain.Lesson)
			for _, i := range enrollIdxs {
				if l, ok := subs[i].Schedule[date]; ok {
					lessons := make([]domain.Lesson, 0, len(l))
					for k := range l {
						lessons = append(lessons, k)
					}
					leftSide[subs[i].SubscriptionUUID] = lessons
				}
			}

			rightSide := make(map[domain.Lesson]struct{})
			for _, lessons := range leftSide {
				for _, l := range lessons {
					rightSide[l] = struct{}{}
				}
			}

			graph, uuidMap, lessonMap := buildGraph(leftSide, rightSide)
			_ = mcmf(graph, 0, len(graph)-1)

			uuidLessons := make(map[uuid.UUID][]domain.Lesson)
			for v, edges := range graph {
				if lesson, ok := lessonMap[v]; ok {
					for _, edge := range edges {
						if edge.Cap == 0 {
							if id, ok := uuidMap[edge.To]; ok {
								uuidLessons[id] = append(uuidLessons[id], lesson)
							}
						}
					}
				}
			}

			for id, lessons := range uuidLessons {
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
			continue
		}
		sub, err := uc.SubscriptionSvc.GetSubscription(ctx, task.sub.SubscriptionUUID)
		if err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: get subscription: %w", err)
			continue
		}
		if sub.Status != subscription.StatusActive {
			continue
		}

		creds, err := uc.AuthSvc.GetUserCredentials(ctx, task.sub.UserUUID)
		if err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: get user credentials: %w", err)
			continue
		}
		if creds.Session == nil || creds.Cookies == nil {
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
			continue
		}

		selectedDate, selectedLesson, ok := pickRandomSlot(freshSchedule)
		if !ok {
			continue
		}

		lessonTime := domain.LessonLookup[int(selectedLesson)]
		targetTime := time.Date(selectedDate.Year(), selectedDate.Month(), selectedDate.Day(), lessonTime.Start.Hour(), lessonTime.Start.Minute(), 0, 0, time.UTC)

		bId, err := uc.EventSvc.Enroll(ctx, &event.EnrollmentReq{
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
			errCh <- fmt.Errorf("event usecase: user worker: enroll: %w", err)
			continue
		}

		fmt.Println(bId)

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

type Edge struct {
	To   int
	Cap  int
	Cost int
	Rev  int
}

func addEdge(graph [][]Edge, from, to, cap, cost int) {
	graph[from] = append(graph[from], Edge{
		To:   to,
		Cap:  cap,
		Cost: cost,
		Rev:  len(graph[to]),
	})
	graph[to] = append(graph[to], Edge{
		To:   from,
		Cap:  0,
		Cost: -cost,
		Rev:  len(graph[from]) - 1,
	})
}

func spfa(graph [][]Edge, source int) ([]int, []int, []int) {
	dist := make([]int, len(graph))
	for i := range graph {
		dist[i] = math.MaxInt32
	}
	dist[source] = 0

	prevv := make([]int, len(graph))
	preve := make([]int, len(graph))
	queue := []int{source}
	inQueue := make([]bool, len(graph))
	inQueue[source] = true

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		inQueue[u] = false

		for i, edge := range graph[u] {
			if edge.Cap > 0 && dist[u]+edge.Cost < dist[edge.To] {
				dist[edge.To] = dist[u] + edge.Cost
				prevv[edge.To] = u
				preve[edge.To] = i
				if !inQueue[edge.To] {
					queue = append(queue, edge.To)
					inQueue[edge.To] = true
				}
			}
		}
	}

	return dist, prevv, preve
}

func augment(graph [][]Edge, source, sink int, dist, prevv, preve []int) int {
	v := sink
	minCap := math.MaxInt32
	for v != source {
		if e := graph[prevv[v]][preve[v]]; e.Cap < minCap {
			minCap = e.Cap
		}
		v = prevv[v]
	}
	v = sink
	for v != source {
		graph[prevv[v]][preve[v]].Cap -= minCap
		e := graph[prevv[v]][preve[v]]
		graph[e.To][e.Rev].Cap += minCap
		v = prevv[v]
	}
	return dist[sink] * minCap
}

func mcmf(graph [][]Edge, source, sink int) int {
	total := 0
	for {
		dist, prevv, preve := spfa(graph, source)
		if dist[sink] == math.MaxInt32 {
			return total
		}
		total += augment(graph, source, sink, dist, prevv, preve)
	}
}

func buildGraph(constraints map[uuid.UUID][]domain.Lesson, rightSide map[domain.Lesson]struct{}) ([][]Edge, map[int]uuid.UUID, map[int]domain.Lesson) {
	uuidMap := make(map[int]uuid.UUID, len(constraints))
	lessonMap := make(map[int]domain.Lesson, len(rightSide))
	reverseUUIDMap := make(map[uuid.UUID]int, len(constraints))
	reverseLessonMap := make(map[domain.Lesson]int, len(rightSide))
	maxLessons := (len(rightSide) + len(constraints) - 1) / len(constraints)

	idx := 1
	for lesson := range rightSide {
		lessonMap[idx] = lesson
		reverseLessonMap[lesson] = idx
		idx++
	}
	for id := range constraints {
		uuidMap[idx] = id
		reverseUUIDMap[id] = idx
		idx++
	}

	graph := make([][]Edge, len(constraints)+len(rightSide)+2)
	for idx := range lessonMap {
		addEdge(graph, 0, idx, 1, 0)
	}
	for idx := range uuidMap {
		addEdge(graph, idx, len(graph)-1, 1, 0)
		addEdge(graph, idx, len(graph)-1, maxLessons-1, 1)
	}
	for id, lessons := range constraints {
		for _, lesson := range lessons {
			addEdge(graph, reverseLessonMap[lesson], reverseUUIDMap[id], 1, 0)
		}
	}

	return graph, uuidMap, lessonMap
}
