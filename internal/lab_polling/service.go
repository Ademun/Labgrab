package lab_polling

import (
	"context"
	"labgrab/internal/shared/api/dikidi"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("lab-polling-service")

type Service struct {
	dikidiClient *dikidi.Client
	slotParser   *Parser
	logger       *zap.SugaredLogger
}

func NewService(client *dikidi.Client, slotParser *Parser, logger *zap.SugaredLogger) *Service {
	return &Service{
		dikidiClient: client,
		slotParser:   slotParser,
		logger:       logger,
	}
}

func (s *Service) GetLabEventsStream(ctx context.Context) chan *Event {
	ctx, span := tracer.Start(ctx, "lab_polling.service.GetLabEventsStream")

	s.logger.Info("starting lab events stream")

	events := make(chan *Event)
	slots := s.dikidiClient.GetSlotStream(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer func() {
			close(events)
			span.End()
			wg.Done()
			s.logger.Info("lab events stream closed")
		}()

		eventCount := 0
		slotCount := 0
		errorCount := 0

		for slot := range slots {
			slotCount++

			if slot.Err != nil {
				errorCount++
				span.RecordError(slot.Err)
				s.logger.Errorw("error receiving slot from dikidi client",
					"error", slot.Err,
					"slot_count", slotCount,
					"error_count", errorCount)
				continue
			}

			parsed, err := s.slotParser.ParseSlot(slot.Data)
			if err != nil {
				errorCount++
				span.RecordError(err)
				s.logger.Errorw("error parsing slot",
					"error", err,
					"slot_count", slotCount,
					"error_count", errorCount)
				continue
			}

			for _, event := range parsed {
				select {
				case events <- &event:
					eventCount++
					if eventCount%100 == 0 {
						s.logger.Infow("lab events stream progress",
							"events_sent", eventCount,
							"slots_processed", slotCount,
							"errors", errorCount)
					}
				case <-ctx.Done():
					span.SetStatus(codes.Error, "context cancelled")
					s.logger.Warnw("lab events stream cancelled by context",
						"events_sent", eventCount,
						"slots_processed", slotCount,
						"errors", errorCount,
						"context_error", ctx.Err())
					return
				}
			}
		}

		span.SetAttributes(
			attribute.Int("events.total", eventCount),
			attribute.Int("slots.total", slotCount),
			attribute.Int("errors.total", errorCount),
		)

		s.logger.Infow("lab events stream completed",
			"events_sent", eventCount,
			"slots_processed", slotCount,
			"errors", errorCount)
	}()

	return events
}
