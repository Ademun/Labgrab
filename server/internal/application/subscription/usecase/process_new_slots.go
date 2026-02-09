package usecase

import (
	"context"
	"labgrab/internal/lab_polling"
	"labgrab/internal/subscription"
	"labgrab/internal/telegram"
	"labgrab/internal/user"
	"sync"

	"go.uber.org/zap"
)

type ProcessNewSlotsUseCase struct {
	labPollingSvc   *lab_polling.Service
	subscriptionSvc *subscription.Service
	userSvc         *user.Service
	telegramSvc     *telegram.Service
	logger          *zap.SugaredLogger
}

func NewProcessNewSlotsUseCase(labPollingSvc *lab_polling.Service, subscriptionSvc *subscription.Service, userSvc *user.Service, telegramSvc *telegram.Service, logger *zap.SugaredLogger) *ProcessNewSlotsUseCase {
	return &ProcessNewSlotsUseCase{
		labPollingSvc:   labPollingSvc,
		subscriptionSvc: subscriptionSvc,
		userSvc:         userSvc,
		telegramSvc:     telegramSvc,
		logger:          logger,
	}
}

func (uc *ProcessNewSlotsUseCase) Exec(ctx context.Context) error {
	currentEvents := uc.labPollingSvc.GetLabEventsStream(ctx)
	sem := make(chan struct{}, 50)
	wg := sync.WaitGroup{}
	totalEvents, matchedSubscriptions := 0, 0
	go func() {
		for event := range currentEvents {
			totalEvents++
			wg.Add(1)
			go func() {
				defer func() {
					<-sem
					wg.Done()
				}()
				sem <- struct{}{}
				err := uc.HandleEvent(ctx, event)
				if err != nil {
					uc.logger.Errorw("error handling event", "event", event, "err", err)
				}
			}()
		}
		wg.Wait()
		close(sem)
	}()
	uc.logger.Infow("Processing complete", "total events", totalEvents, "matched subscriptions", matchedSubscriptions)

	return nil
}

func (uc *ProcessNewSlotsUseCase) HandleEvent(ctx context.Context, event *lab_polling.Event) error {
	searchReq := &subscription.GetMatchingSubscriptionsReq{
		LabType:        event.Type,
		LabTopic:       event.Topic,
		LabNumber:      event.Number,
		LabAuditorium:  event.Auditorium,
		AvailableSlots: event.Schedule,
	}

	relevantSubs, err := uc.subscriptionSvc.GetMatchingSubscriptions(ctx, searchReq)
	if err != nil {
		return err
	}

	for _, sub := range relevantSubs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			userData, err := uc.userSvc.GetUser(ctx, sub.UserUUID)
			if err != nil {
				return err
			}

			notifyReq := telegram.NotifyUserReq{
				UserID:        userData.TelegramID,
				LabName:       event.Name,
				LabType:       string(event.Type),
				LabTopic:      string(event.Topic),
				LabNumber:     event.Number,
				LabAuditorium: event.Auditorium,
				Schedule:      sub.MatchingTimeslots,
				PageURL:       "stub", // TODO: include
			}

			return uc.telegramSvc.NotifyUser(ctx, notifyReq)
		}
	}

	return nil
}
