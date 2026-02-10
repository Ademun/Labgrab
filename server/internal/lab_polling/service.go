package lab_polling

import (
	"context"
	"labgrab/internal/shared/api/dikidi"
	"sync"

	"go.opentelemetry.io/otel"
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

func (s *Service) GetLabEventsStream(ctx context.Context) chan *EventResult {
	events := make(chan *EventResult)
	slots := s.dikidiClient.GetSlotStream(ctx)

	go func() {
		wg := &sync.WaitGroup{}

		for slot := range slots {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if slot.Err != nil {
					events <- &EventResult{
						Data: nil,
						Err:  slot.Err,
					}
				}

				parsed, err := s.slotParser.ParseSlot(slot.Data)
				if err != nil {
					events <- &EventResult{
						Data: nil,
						Err:  err,
					}
				}

				for _, event := range parsed {
					select {
					case events <- &EventResult{
						Data: &event,
						Err:  nil,
					}:
					case <-ctx.Done():
						return
					}
				}
			}()
		}

		wg.Wait()
		close(events)
	}()

	return events
}
