package usecase

import (
	"context"
	"errors"
	"fmt"
	"labgrab/internal/booking"
	"labgrab/internal/event"
	"labgrab/internal/shared/domain"
	"labgrab/internal/subscription"
	"labgrab/internal/telegram"
	"labgrab/internal/user"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ProcessEventsUsecase struct {
	eventSvc        *event.Service
	userSvc         *user.Service
	bookingSvc      *booking.Service
	subscriptionSvc *subscription.Service
	telegramSvc     *telegram.Service
	tracer          trace.Tracer
}

func NewProcessEventsUsecase(
	eventSvc *event.Service,
	userSvc *user.Service,
	bookingSvc *booking.Service,
	subscriptionSvc *subscription.Service,
	telegramSvc *telegram.Service,
) *ProcessEventsUsecase {
	return &ProcessEventsUsecase{
		eventSvc:        eventSvc,
		userSvc:         userSvc,
		bookingSvc:      bookingSvc,
		subscriptionSvc: subscriptionSvc,
		telegramSvc:     telegramSvc,
		tracer:          otel.Tracer("process_events_usecase"),
	}
}

func (uc *ProcessEventsUsecase) Exec(ctx context.Context) error {
	ctx, span := uc.tracer.Start(ctx, "process_events_usecase.Exec")
	defer span.End()

	var clientCookies string
	events, err := uc.eventSvc.GetCurrentEvents(ctx, &clientCookies)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to process events")
		return fmt.Errorf("event: usecase: process events: %w", err)
	}

	wg := sync.WaitGroup{}
	errCh := make(chan error)

	go func() {
		for range 5 {
			wg.Add(1)
			go func() {
				uc.Worker(ctx, events, errCh)
				wg.Done()
			}()
		}
		wg.Wait()
		close(errCh)
	}()

	var collected error
	for err := range errCh {
		fmt.Println("Error", err)
		errors.Join(collected, err)
	}

	if collected != nil {
		span.RecordError(collected)
		span.SetStatus(codes.Error, "failed to process events")
		return fmt.Errorf("event: usecase: process events: %w", collected)
	}

	return nil
}

func (uc *ProcessEventsUsecase) Worker(ctx context.Context, events chan *event.GetEventsRes, errCh chan error) {
main:
	for e := range events {
		if e.Err != nil {
			errCh <- fmt.Errorf("event: worker: event error: %w", e.Err)
			continue main
		}

		subs, err := uc.subscriptionSvc.GetMatchingSubscriptions(ctx, &subscription.GetMatchingSubscriptionsReq{
			Type:       e.Data.Type,
			Topic:      e.Data.Topic,
			Number:     e.Data.Number,
			Auditorium: e.Data.Auditorium,
			Schedule:   e.Data.Schedule,
		})
		if err != nil {
			errCh <- fmt.Errorf("event: worker: match subscriptions: %w", err)
			continue main
		}

		enrollmentSubscriptions := make([]subscription.GetMatchingSubscriptionsRes, 0)
		subUserInfoIdx := make(map[uuid.UUID]*user.GetUserRes, len(subs))

		for i := range subs {
			userInfo, err := uc.userSvc.GetUser(ctx, subs[i].UserUUID)
			if err != nil {
				errCh <- fmt.Errorf("event: worker: get user: %w", err)
				continue main
			}
			subUserInfoIdx[subs[i].SubscriptionUUID] = userInfo
			if !subs[i].AutoEnroll {
				fmt.Println(subs[i], "notify")
				err := uc.telegramSvc.NotifyUser(ctx, telegram.NotifyUserReq{
					UserID:        userInfo.TelegramID,
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
					errCh <- fmt.Errorf("event: worker: notify: %w", err)
					continue main
				}
			} else {
				enrollmentSubscriptions = append(enrollmentSubscriptions, subs[i])
			}
		}

		keyMap := make(map[time.Time]int)
		conflictingDates := make([]time.Time, 0)
		for i := range enrollmentSubscriptions {
			filteredSchedule, err := uc.bookingSvc.FilterSchedule(ctx, &booking.FilterScheduleReq{
				UserUUID: subs[i].UserUUID,
				Schedule: subs[i].Schedule,
			})
			if err != nil {
				errCh <- fmt.Errorf("event: worker: filter schedule: %w", err)
				continue main
			}

			subs[i].Schedule = filteredSchedule

			for d := range filteredSchedule {
				utcD := d.UTC()
				keyMap[utcD]++
				if keyMap[utcD] == 2 {
					conflictingDates = append(conflictingDates, d)
				}
			}
		}

		subIndex := make(map[uuid.UUID]int, len(enrollmentSubscriptions))
		for i := range enrollmentSubscriptions {
			subIndex[subs[i].SubscriptionUUID] = i
		}

		for _, date := range conflictingDates {
			leftSide := make(map[uuid.UUID][]domain.Lesson)
			for i := range subs {
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
			for v, e := range graph {
				if lesson, ok := lessonMap[v]; ok {
					for _, edge := range e {
						if edge.Cap == 0 {
							id, ok := uuidMap[edge.To]
							if ok {
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

		for i := range enrollmentSubscriptions {
			userInfo := subUserInfoIdx[enrollmentSubscriptions[i].SubscriptionUUID]
			//fmt.Println("info", userInfo, "sub", enrollmentSubscriptions[i], "event", e.Data)
			fmt.Println("User", userInfo.Name, "Estimated schedule", enrollmentSubscriptions[i].Schedule)
		}
	}
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

	queue := make([]int, 0)
	inQueue := make([]bool, len(graph))

	queue = append(queue, source)
	inQueue[source] = true

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		inQueue[u] = false

		for i, edge := range graph[u] {
			if (dist[u]+edge.Cost < dist[edge.To]) && (edge.Cap > 0) {
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
		edge := graph[prevv[v]][preve[v]]
		if edge.Cap < minCap {
			minCap = edge.Cap
		}
		v = prevv[v]
	}
	v = sink
	for v != source {
		graph[prevv[v]][preve[v]].Cap -= minCap
		edge := graph[prevv[v]][preve[v]]
		graph[edge.To][edge.Rev].Cap += minCap
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
