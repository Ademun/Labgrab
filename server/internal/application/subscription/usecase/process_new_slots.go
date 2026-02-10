package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/lab_polling"
	"labgrab/internal/subscription"
	"labgrab/internal/telegram"
	"labgrab/internal/user"
	"sync"

	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type ProcessNewSlotsUseCase struct {
	labPollingSvc   *lab_polling.Service
	subscriptionSvc *subscription.Service
	userSvc         *user.Service
	telegramSvc     *telegram.Service
}

func NewProcessNewSlotsUseCase(labPollingSvc *lab_polling.Service, subscriptionSvc *subscription.Service, userSvc *user.Service, telegramSvc *telegram.Service, logger *zap.SugaredLogger) *ProcessNewSlotsUseCase {
	return &ProcessNewSlotsUseCase{
		labPollingSvc:   labPollingSvc,
		subscriptionSvc: subscriptionSvc,
		userSvc:         userSvc,
		telegramSvc:     telegramSvc,
	}
}

func (uc *ProcessNewSlotsUseCase) Exec(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Exec")
	defer span.End()

	currentEvents := uc.labPollingSvc.GetLabEventsStream(ctx)

	sem := make(chan struct{}, 50)
	eventCount := 0
	errs := make([]error, 0)

	mainWg := sync.WaitGroup{}
	mainWg.Add(1)

	go func() {
		wg := sync.WaitGroup{}
		for event := range currentEvents {
			eventCount++
			wg.Add(1)
			go func() {
				defer func() {
					<-sem
					wg.Done()
				}()
				sem <- struct{}{}
				err := uc.HandleEvent(ctx, event)
				if err != nil {
					span.RecordError(err)
					errs = append(errs, err)
				}
			}()
		}
		wg.Wait()
		close(sem)
		mainWg.Done()
	}()
	mainWg.Wait()

	if len(errs) > 0 {
		err := fmt.Errorf("processed %d events. %d failed", eventCount, len(errs))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (uc *ProcessNewSlotsUseCase) HandleEvent(ctx context.Context, event *lab_polling.EventResult) error {
	ctx, span := tracer.Start(ctx, "HandleEvent")
	defer span.End()

	if event.Err != nil {
		span.RecordError(event.Err)
		span.SetStatus(codes.Error, event.Err.Error())
		return event.Err
	}

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
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	for _, sub := range relevantSubs {
		select {
		case <-ctx.Done():
			span.RecordError(ctx.Err())
			span.SetStatus(codes.Error, ctx.Err().Error())
			return ctx.Err()
		default:
			userData, err := uc.userSvc.GetUser(ctx, sub.UserUUID)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return err
			}

			notifyReq := telegram.NotifyUserReq{
				UserID:        userData.TelegramID,
				LabName:       event.Data.Name,
				LabType:       string(event.Data.Type),
				LabTopic:      string(event.Data.Topic),
				LabNumber:     event.Data.Number,
				LabAuditorium: event.Data.Auditorium,
				Schedule:      sub.MatchingTimeslots,
				PageURL:       "stub", // TODO: include
			}

			return uc.telegramSvc.NotifyUser(ctx, notifyReq)
		}
	}

	span.SetStatus(codes.Ok, "")

	return nil
}
