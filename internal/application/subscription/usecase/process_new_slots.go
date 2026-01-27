package usecase

import (
	"context"
	"labgrab/internal/lab_polling"
	"labgrab/internal/subscription"
	"sync"

	"go.uber.org/zap"
)

type ProcessNewSlotsUseCase struct {
	labPollingSvc   *lab_polling.Service
	subscriptionSvc *subscription.Service
	logger          *zap.SugaredLogger
}

func NewProcessNewSlotsUseCase(labPollingSvc *lab_polling.Service, subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *ProcessNewSlotsUseCase {
	return &ProcessNewSlotsUseCase{
		labPollingSvc:   labPollingSvc,
		subscriptionSvc: subscriptionSvc,
		logger:          logger,
	}
}

func (uc *ProcessNewSlotsUseCase) Exec(ctx context.Context) error {
	currentEvents := uc.labPollingSvc.GetLabEventsStream(ctx)
	targetSubscriptions := make(chan subscription.GetMatchingSubscriptionsRes)
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
				err := uc.HandleEvent(ctx, event, targetSubscriptions)
				if err != nil {
					uc.logger.Errorw("error handling event", "event", event, "err", err)
				}
			}()
		}
		wg.Wait()
		close(targetSubscriptions)
		close(sem)
	}()

	for sub := range targetSubscriptions {
		matchedSubscriptions++
		uc.logger.Infow("Processing subscription", "subscription", sub)
	}
	uc.logger.Infow("Processing complete", "total events", totalEvents, "matched subscriptions", matchedSubscriptions)

	return nil
}

func (uc *ProcessNewSlotsUseCase) HandleEvent(ctx context.Context, event *lab_polling.Event, subsChan chan subscription.GetMatchingSubscriptionsRes) error {
	searchReq := &subscription.GetMatchingSubscriptionsReq{
		LabType:        subscription.LabType(event.Type),
		LabTopic:       subscription.LabTopic(event.Topic),
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
		case subsChan <- sub:
		}
	}

	return nil
}
