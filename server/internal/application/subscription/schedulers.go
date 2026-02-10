package subscription

import (
	"context"
	"labgrab/internal/application/subscription/usecase"
	"labgrab/internal/lab_polling"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/subscription"
	"labgrab/internal/telegram"
	"labgrab/internal/user"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type Scheduler struct {
	dikidiClient    *dikidi.Client
	logger          *zap.SugaredLogger
	scheduler       gocron.Scheduler
	processNewSlots *usecase.ProcessNewSlotsUseCase
}

func NewScheduler(dikidiClient *dikidi.Client, pollingSvc *lab_polling.Service, subscriptionSvc *subscription.Service, userSvc *user.Service, telegramSvc *telegram.Service, logger *zap.SugaredLogger) *Scheduler {
	return &Scheduler{
		dikidiClient:    dikidiClient,
		logger:          logger,
		processNewSlots: usecase.NewProcessNewSlotsUseCase(pollingSvc, subscriptionSvc, userSvc, telegramSvc, logger),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.UpdateSlotSources(ctx)
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return err
	}
	_, err = scheduler.NewJob(
		gocron.DurationRandomJob(time.Minute, time.Minute*10),
		gocron.NewTask(s.ProcessNewSlots, ctx),
	)
	if err != nil {
		return err
	}
	_, err = scheduler.NewJob(
		gocron.DurationRandomJob(time.Hour, time.Hour*24),
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
	ctx, span := tracer.Start(ctx, "ProcessNewSlots")
	defer span.End()

	err := s.processNewSlots.Exec(ctx)
	if err != nil {
		s.logger.Errorw("Error executing process new slots",
			"trace_id", span.SpanContext().TraceID(),
			"span_id", span.SpanContext().SpanID(),
			"error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.SetStatus(codes.Ok, "")
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
