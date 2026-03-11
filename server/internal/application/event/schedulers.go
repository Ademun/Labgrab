package event

import (
	"context"
	"labgrab/internal/application/event/usecase"
	"labgrab/internal/auth"
	"labgrab/internal/booking"
	"labgrab/internal/event"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/subscription"
	"labgrab/internal/telegram"
	"labgrab/internal/user"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type Scheduler struct {
	dikidiClient     *dikidi.Client
	scheduler        gocron.Scheduler
	processNewEvents *usecase.ProcessEventsUsecase
	updateServiceIDs *usecase.UpdateServiceIDsUsecase
	logger           *zap.SugaredLogger
}

func NewScheduler(
	dikidiClient *dikidi.Client,
	eventSvc *event.Service,
	subscriptionSvc *subscription.Service,
	userSvc *user.Service,
	authSvc *auth.Service,
	bookingSvc *booking.Service,
	telegramSvc *telegram.Service,
	logger *zap.SugaredLogger,
) *Scheduler {
	return &Scheduler{
		dikidiClient: dikidiClient,
		logger:       logger,
		processNewEvents: &usecase.ProcessEventsUsecase{
			EventSvc:        eventSvc,
			UserSvc:         userSvc,
			AuthSvc:         authSvc,
			BookingSvc:      bookingSvc,
			SubscriptionSvc: subscriptionSvc,
			TelegramSvc:     telegramSvc,
		},
		updateServiceIDs: &usecase.UpdateServiceIDsUsecase{
			EventSvc: eventSvc,
		},
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.UpdateServiceIDs(ctx)
	s.ProcessNewEvents(ctx)
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	_, err = scheduler.NewJob(
		gocron.DurationRandomJob(time.Minute*10, time.Minute*30),
		gocron.NewTask(s.ProcessNewEvents, ctx),
	)
	if err != nil {
		return err
	}

	_, err = scheduler.NewJob(
		gocron.DurationRandomJob(time.Hour, time.Hour*24),
		gocron.NewTask(s.UpdateServiceIDs, ctx),
	)
	if err != nil {
		return err
	}

	scheduler.Start()
	s.scheduler = scheduler
	return nil
}

func (s *Scheduler) Stop() error {
	return s.scheduler.Shutdown()
}

func (s *Scheduler) ProcessNewEvents(ctx context.Context) {
	if err := s.processNewEvents.Exec(ctx); err != nil {
		s.logger.Errorf("event scheduler: process new events: %v", err)
		return
	}
}

func (s *Scheduler) UpdateServiceIDs(ctx context.Context) {
	if err := s.updateServiceIDs.Exec(ctx); err != nil {
		s.logger.Errorf("event scheduler: update service ids: %v", err)
		return
	}
}
