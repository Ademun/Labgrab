package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/event"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type UpdateServiceIDsUsecase struct {
	eventSvc *event.Service
	tracer   trace.Tracer
}

func NewUpdateServiceIDsUsecase(
	eventSvc *event.Service,
) *UpdateServiceIDsUsecase {
	return &UpdateServiceIDsUsecase{
		eventSvc: eventSvc,
		tracer:   otel.Tracer("update_service_ids_usecase"),
	}
}

func (uc *UpdateServiceIDsUsecase) Exec(ctx context.Context) error {
	ctx, span := uc.tracer.Start(ctx, "update_service_ids_usecase.Exec")
	defer span.End()

	var clientCookies string
	if err := uc.eventSvc.UpdateServiceIDs(ctx, &clientCookies); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update service ids")
		return fmt.Errorf("event: usecase: update service ids: %w", err)
	}

	return nil
}
