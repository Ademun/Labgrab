package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/lab_polling"
	"labgrab/internal/record"
	"labgrab/internal/subscription"
	"labgrab/internal/telegram"
	"labgrab/internal/user"
	"log/slog"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type ProcessNewSlotsUseCase struct {
	labPollingSvc   *lab_polling.Service
	subscriptionSvc *subscription.Service
	userSvc         *user.Service
	recordSvc       *record.Service
	telegramSvc     *telegram.Service
	logger          *zap.SugaredLogger
}

func NewProcessNewSlotsUseCase(labPollingSvc *lab_polling.Service, subscriptionSvc *subscription.Service, userSvc *user.Service, recordSvc *record.Service, telegramSvc *telegram.Service, logger *zap.SugaredLogger) *ProcessNewSlotsUseCase {
	return &ProcessNewSlotsUseCase{
		labPollingSvc:   labPollingSvc,
		subscriptionSvc: subscriptionSvc,
		userSvc:         userSvc,
		recordSvc:       recordSvc,
		telegramSvc:     telegramSvc,
		logger:          logger,
	}
}

func (uc *ProcessNewSlotsUseCase) Exec(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "subscription_usecase.process_new_slots")
	defer span.End()

	currentEvents := uc.labPollingSvc.GetLabEventsStream(ctx)

	sem := make(chan struct{}, 50)
	var eventCount atomic.Int64
	var errorCount atomic.Int64
	errs := make([]error, 0)
	errMux := sync.Mutex{}

	mainWg := sync.WaitGroup{}
	mainWg.Add(1)

	go func() {
		wg := sync.WaitGroup{}
		for event := range currentEvents {
			eventCount.Add(1)
			wg.Add(1)
			go func(evt *lab_polling.EventResult) {
				defer func() {
					<-sem
					wg.Done()
				}()
				sem <- struct{}{}

				err := uc.HandleEvent(ctx, evt)
				if err != nil {
					errorCount.Add(1)
					span.RecordError(err)

					errMux.Lock()
					errs = append(errs, err)
					errMux.Unlock()
				}
			}(event)
		}
		wg.Wait()
		close(sem)
		mainWg.Done()
	}()
	mainWg.Wait()

	totalEvents := eventCount.Load()
	totalErrors := errorCount.Load()

	span.SetAttributes(
		attribute.Int64("events.total", totalEvents),
		attribute.Int64("events.successful", totalEvents-totalErrors),
		attribute.Int64("events.failed", totalErrors),
	)

	if len(errs) > 0 {
		err := fmt.Errorf("processed %d events. %d failed", totalEvents, len(errs))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (uc *ProcessNewSlotsUseCase) HandleEvent(ctx context.Context, event *lab_polling.EventResult) error {
	ctx, span := tracer.Start(ctx, "subscription_usecase.handle_event")
	defer span.End()

	if event.Err != nil {
		span.RecordError(event.Err)
		span.SetStatus(codes.Error, "event contains error")
		return event.Err
	}

	span.SetAttributes(
		attribute.String("lab.name", event.Data.Name),
		attribute.String("lab.type", string(event.Data.Type)),
		attribute.String("lab.topic", string(event.Data.Topic)),
		attribute.Int("lab.number", event.Data.Number),
		attribute.Int("lab.auditorium", event.Data.Auditorium),
	)

	searchReq := &subscription.GetMatchingSubscriptionsReq{
		LabType:        event.Data.Type,
		LabTopic:       event.Data.Topic,
		LabNumber:      event.Data.Number,
		LabAuditorium:  event.Data.Auditorium,
		AvailableSlots: event.Data.Schedule,
	}

	relevantSubs, err := uc.subscriptionSvc.GetMatchingSubscriptions(ctx, searchReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get matching subscriptions")
		return err
	}

	span.SetAttributes(attribute.Int("subscriptions.matched", len(relevantSubs)))

	for _, sub := range relevantSubs {
		select {
		case <-ctx.Done():
			span.RecordError(ctx.Err())
			span.SetStatus(codes.Error, "context cancelled")
			return ctx.Err()
		default:
			userData, err := uc.userSvc.GetUser(ctx, sub.UserUUID)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to retrieve user data")
				span.End()
				return err
			}

			fTimeslots, err := uc.recordSvc.FilterAvailableSlots(ctx, &record.FilterSlotsReq{
				UserUUID:          sub.UserUUID,
				MatchingTimeslots: sub.MatchingTimeslots,
			})

			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to filter available slots")
				span.End()
				return err
			}

			sub.MatchingTimeslots = fTimeslots

			slog.Info("slot", sub)

			if !sub.AutoEnroll {
				span.SetAttributes(attribute.Int64("telegram.user_id", int64(userData.TelegramID)))

				err = uc.telegramSvc.NotifyUser(ctx, telegram.NotifyUserReq{
					UserID:        userData.TelegramID,
					LabName:       event.Data.Name,
					LabType:       event.Data.Type,
					LabTopic:      event.Data.Topic,
					LabNumber:     event.Data.Number,
					LabAuditorium: event.Data.Auditorium,
					Spot:          event.Data.Spot,
					Schedule:      sub.MatchingTimeslots,
					PageURL:       "https://dikidi.net/550001?p=0.pi-ssm-sd&s=5300027&rl=0_undefined", // TODO: include
				})
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, "failed to send telegram notification")
					span.End()
					return err
				}
			}

		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
