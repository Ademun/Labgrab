package subscription

import (
	"context"
	"labgrab/internal/application/subscription/usecase"
	"labgrab/internal/lab_polling"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/subscription"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type Scheduler struct {
	dikidiClient    *dikidi.Client
	pollingSvc      *lab_polling.Service
	subscriptionSvc *subscription.Service
	logger          *zap.SugaredLogger
	scheduler       gocron.Scheduler
	processNewSlots *usecase.ProcessNewSlotsUseCase
}

func NewScheduler(dikidiClient *dikidi.Client, pollingSvc *lab_polling.Service, subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *Scheduler {
	return &Scheduler{
		dikidiClient:    dikidiClient,
		pollingSvc:      pollingSvc,
		subscriptionSvc: subscriptionSvc,
		logger:          logger,
		processNewSlots: usecase.NewProcessNewSlotsUseCase(pollingSvc, subscriptionSvc, logger),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.UpdateSlotSources(ctx)
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return err
	}
	_, err = scheduler.NewJob(
		gocron.DurationRandomJob(time.Second*30, time.Minute),
		gocron.NewTask(s.ProcessNewSlots, ctx),
	)
	if err != nil {
		return err
	}
	_, err = scheduler.NewJob(
		gocron.DurationRandomJob(time.Hour*12, time.Hour*24),
		gocron.NewTask(s.UpdateSlotSources, ctx),
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

func (s *Scheduler) ProcessNewSlots(ctx context.Context) {
	now := time.Now()
	s.logger.Infow("Running job", "job", "ProcessNewSlots", "time", now)
	err := s.processNewSlots.Exec(ctx)
	if err != nil {
		s.logger.Errorw("Error executing process new slots", "error", err)
	}
	s.logger.Infow("Finished running job", "job", "ProcessNewSlots", "elapsed", time.Now().Sub(now))
}

func (s *Scheduler) UpdateSlotSources(ctx context.Context) {
	now := time.Now()
	s.logger.Infow("Running job", "job", "UpdateSlotSources", "time", now)
	err := s.dikidiClient.UpdateSlotSourceIDs(ctx)
	if err != nil {
		s.logger.Errorw("Error updating slot sources", "error", err)
	}
	s.logger.Infow("Finished running job", "job", "UpdateSlotSources", "elapsed", time.Now().Sub(now))
}
