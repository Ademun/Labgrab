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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// maxConcurrentRequests matches the site's concurrency limit.
const maxConcurrentRequests = 5

type ProcessEventsUsecase struct {
	eventSvc        *event.Service
	userSvc         *user.Service
	authSvc         *auth.Service
	bookingSvc      *booking.Service
	subscriptionSvc *subscription.Service
	telegramSvc     *telegram.Service
	tracer          trace.Tracer
}

func NewProcessEventsUsecase(
	eventSvc *event.Service,
	userSvc *user.Service,
	authSvc *auth.Service,
	bookingSvc *booking.Service,
	subscriptionSvc *subscription.Service,
	telegramSvc *telegram.Service,
) *ProcessEventsUsecase {
	return &ProcessEventsUsecase{
		eventSvc:        eventSvc,
		userSvc:         userSvc,
		authSvc:         authSvc,
		bookingSvc:      bookingSvc,
		subscriptionSvc: subscriptionSvc,
		telegramSvc:     telegramSvc,
		tracer:          otel.Tracer("process_events_usecase"),
	}
}

// enrollTask is produced by event workers after MCMF conflict resolution
// and consumed by per-user workers for the actual enrollment.
type enrollTask struct {
	eventRes *event.GetEventsRes
	sub      subscription.GetMatchingSubscriptionsRes
	userInfo *user.GetUserRes
}

// Exec orchestrates two-phase parallel processing:
//
//	Phase 1 — 3 event workers: match subscriptions, filter schedule, resolve
//	          inter-user conflicts via MCMF, fan-out enrollTasks.
//	Phase 2 — per-user goroutines: re-filter (fresh DB), enroll (semaphore),
//	          update state, notify, fetch updated bookings (semaphore).
//
// One global semaphore caps all outbound HTTP at maxConcurrentRequests.
// Per-user serialization guarantees the 2-day gap invariant without any mutex.
func (uc *ProcessEventsUsecase) Exec(ctx context.Context) error {
	ctx, span := uc.tracer.Start(ctx, "process_events_usecase.Exec")
	defer span.End()

	var clientCookies string
	events, err := uc.eventSvc.GetCurrentEvents(ctx, &clientCookies)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get current events")
		return fmt.Errorf("event usecase: exec: get current events: %w", err)
	}

	// Shared HTTP semaphore — every outbound request to the site acquires one slot.
	sem := make(chan struct{}, maxConcurrentRequests)

	// enrollCh is buffered so event workers never block on slow user workers.
	enrollCh := make(chan enrollTask, 100)
	errCh := make(chan error, 100)

	// ── Phase 1: event workers ───────────────────────────────────────────────
	var eventWg sync.WaitGroup
	for range 3 {
		eventWg.Add(1)
		go func() {
			defer eventWg.Done()
			uc.eventWorker(ctx, events, enrollCh, errCh)
		}()
	}

	// Close enrollCh once all event workers exit — signals fan-out to stop.
	go func() {
		eventWg.Wait()
		close(enrollCh)
	}()

	// ── Phase 2: fan-out + per-user goroutines ───────────────────────────────
	// Single goroutine owns userChannels — no mutex needed.
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
				uc.userWorker(ctx, userCh, sem, errCh)
			}(ch)
		}
		userChannels[userID] <- task
	}

	// enrollCh drained → all tasks routed → close per-user channels so workers exit.
	for _, ch := range userChannels {
		close(ch)
	}

	userWg.Wait()
	close(errCh)

	// Collect all errors. Fixed: errors.Join result must be assigned.
	var collected error
	for err := range errCh {
		fmt.Println(err)
		collected = errors.Join(collected, err)
	}

	if collected != nil {
		span.RecordError(collected)
		span.SetStatus(codes.Error, "failed to process events")
		return fmt.Errorf("event usecase: exec: %w", collected)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// eventWorker reads events, matches subscriptions, runs MCMF conflict resolution,
// and fans out enrollTasks. Pure computation + DB reads — no HTTP, no semaphore.
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

		subs, err := uc.subscriptionSvc.GetMatchingSubscriptions(ctx, &subscription.GetMatchingSubscriptionsReq{
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

		// Pre-fetch all users so we have them for both notify and enroll paths.
		subUsers := make([]*user.GetUserRes, len(subs))
		for i := range subs {
			userInfo, err := uc.userSvc.GetUser(ctx, subs[i].UserUUID)
			if err != nil {
				errCh <- fmt.Errorf("event usecase: event worker: get user: %w", err)
				continue
			}
			subUsers[i] = userInfo
		}

		// Split into notify-only and auto-enroll paths.
		enrollIdxs := make([]int, 0, len(subs))
		for i := range subs {
			if subUsers[i] == nil {
				continue
			}

			if !subs[i].AutoEnroll {
				err := uc.telegramSvc.NotifyEvent(ctx, telegram.NotifyEventReq{
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

			// First-pass schedule filter — used as MCMF input.
			// User worker re-filters with fresh state before actual enrollment.
			filteredSchedule, err := uc.bookingSvc.FilterSchedule(ctx, &booking.FilterScheduleReq{
				UserUUID: subs[i].UserUUID,
				Schedule: subs[i].Schedule,
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

		// ── MCMF conflict resolution ─────────────────────────────────────────
		// Find dates where 2+ users want to enroll, resolve fair assignment.
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

		// Fan-out — each sub becomes an independent task routed by userID.
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
	sem chan struct{},
	errCh chan<- error,
) {
	for task := range tasks {
		freshSchedule, err := uc.bookingSvc.FilterSchedule(ctx, &booking.FilterScheduleReq{
			UserUUID: task.sub.UserUUID,
			Schedule: task.sub.Schedule,
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

		if err := acquireSem(ctx, sem); err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: enroll acquire sem: %w", err)
			continue
		}

		releaseSem(sem)

		creds, err := uc.authSvc.GetUserCredentials(ctx, task.sub.UserUUID)
		if err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: get user credentials: %w", err)
			continue
		}

		if creds.Session == nil || creds.Cookies == nil {
			errCh <- fmt.Errorf("event usecase: user worker: get user credentials: %w", err)
			continue
		}

		if err := acquireSem(ctx, sem); err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: load bookings acquire sem: %w", err)
			continue
		}
		bId, err := uc.eventSvc.Enroll(ctx, &event.EnrollmentReq{
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
			errCh <- fmt.Errorf("event usecase: user worker: enroll acquire sem: %w", err)
			return
		}
		releaseSem(sem)

		fmt.Println(bId)

		if err = uc.subscriptionSvc.CloseSubscription(ctx, task.sub.SubscriptionUUID); err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: update subscription: %w", err)
		}

		if err = uc.telegramSvc.NotifyEnrollment(ctx, telegram.NotifyEnrollmentReq{
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

		if err := acquireSem(ctx, sem); err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: load bookings acquire sem: %w", err)
			continue
		}
		err = uc.bookingSvc.LoadClientBookings(ctx, &booking.LoadClientBookingsReq{
			UserUUID: task.sub.UserUUID,
			Session:  *creds.Session,
			Cookies:  *creds.Cookies,
		})
		releaseSem(sem)
		if err != nil {
			errCh <- fmt.Errorf("event usecase: user worker: load client bookings: %w", err)
		}
	}
}

func acquireSem(ctx context.Context, sem chan struct{}) error {
	select {
	case sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func releaseSem(sem chan struct{}) {
	<-sem
}

// pickRandomSlot selects a random date and lesson from a filtered schedule.
// Returns false if the schedule is empty or all dates have no lessons.
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

// ── MCMF ─────────────────────────────────────────────────────────────────────

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
