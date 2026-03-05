package event

import (
	"context"
	"labgrab/internal/application/event/usecase"
	"labgrab/internal/booking"
	"labgrab/internal/event"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/subscription"
	"labgrab/internal/telegram"
	"labgrab/internal/user"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Scheduler struct {
	dikidiClient     *dikidi.Client
	logger           *zap.SugaredLogger
	scheduler        gocron.Scheduler
	processNewSlots  *usecase.ProcessEventsUsecase
	updateServiceIDs *usecase.UpdateServiceIDsUsecase
	tracer           trace.Tracer
}

func NewScheduler(
	dikidiClient *dikidi.Client,
	eventSvc *event.Service,
	subscriptionSvc *subscription.Service,
	userSvc *user.Service,
	bookingSvc *booking.Service,
	telegramSvc *telegram.Service,
	logger *zap.SugaredLogger,
) *Scheduler {
	return &Scheduler{
		dikidiClient: dikidiClient,
		logger:       logger,
		processNewSlots: usecase.NewProcessEventsUsecase(
			eventSvc,
			userSvc,
			bookingSvc,
			subscriptionSvc,
			telegramSvc,
		),
		updateServiceIDs: usecase.NewUpdateServiceIDsUsecase(
			eventSvc,
		),
		tracer: otel.Tracer("event_scheduler"),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.UpdateServiceIDs(ctx)
	s.processNewSlots.Exec(ctx)
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return err
	}
	_, err = scheduler.NewJob(
		gocron.DurationRandomJob(time.Hour*5, time.Hour*20),
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
	ctx, span := s.tracer.Start(ctx, "event_scheduler.ProcessNewEvents")
	defer span.End()

	err := s.processNewSlots.Exec(ctx)
	if err != nil {
		s.logger.Errorw("Failed to process new events",
			"trace_id", span.SpanContext().TraceID(),
			"span_id", span.SpanContext().SpanID(),
			"error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to process new events")
		return
	}

	span.SetStatus(codes.Ok, "")
}

func (s *Scheduler) UpdateServiceIDs(ctx context.Context) {
	ctx, span := s.tracer.Start(ctx, "event_scheduler.UpdateServiceIDs")
	defer span.End()

	if err := s.updateServiceIDs.Exec(ctx); err != nil {
		s.logger.Errorw("Failed to update service IDs",
			"trace_id", span.SpanContext().TraceID(),
			"span_id", span.SpanContext().SpanID(),
			"error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update service IDs")
		return
	}

	span.SetStatus(codes.Ok, "")
}
