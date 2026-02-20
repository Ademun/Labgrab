package usecase

import (
	"context"
	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/auth"
	"labgrab/internal/shared/domain"
	"labgrab/internal/subscription"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type NewSubscriptionUseCase struct {
	authSvc         *auth.Service
	subscriptionSvc *subscription.Service
}

func NewNewSubscriptionUseCase(authSvc *auth.Service, subscriptionSvc *subscription.Service) *NewSubscriptionUseCase {
	return &NewSubscriptionUseCase{
		authSvc:         authSvc,
		subscriptionSvc: subscriptionSvc,
	}
}

func (uc *NewSubscriptionUseCase) Exec(ctx context.Context, session string, data *dto.NewSubscriptionReqDTO) (*dto.NewSubscriptionResDTO, error) {
	ctx, span := tracer.Start(ctx, "subscription_usecase.new_subscription")
	defer span.End()

	if err := uc.authSvc.ValidateSession(ctx, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "session validation failed")
		return nil, err
	}

	userUUID, err := uc.authSvc.GetSessionData(ctx, session)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve session data")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.uuid", userUUID.String()),
		attribute.String("lab.type", data.LabType),
		attribute.String("lab.topic", data.LabTopic),
		attribute.Int("lab.number", data.LabNumber),
	)

	if data.LabAuditorium != nil {
		span.SetAttributes(attribute.Int("lab.auditorium", *data.LabAuditorium))
	}

	req := &subscription.CreateSubscriptionReq{
		UserUUID:      userUUID,
		LabType:       domain.LabType(data.LabType),
		LabTopic:      domain.LabTopic(data.LabTopic),
		LabNumber:     data.LabNumber,
		LabAuditorium: data.LabAuditorium,
		CreatedAt:     time.Now(),
	}

	subscriptionUUID, err := uc.subscriptionSvc.CreateSubscription(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create subscription")
		return nil, err
	}

	span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID.String()))
	span.SetStatus(codes.Ok, "")

	return &dto.NewSubscriptionResDTO{
		UUID: subscriptionUUID.String(),
	}, nil
}
