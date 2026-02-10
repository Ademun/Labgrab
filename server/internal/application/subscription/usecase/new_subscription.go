package usecase

import (
	"context"
	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/auth"
	"labgrab/internal/shared/domain"
	"labgrab/internal/subscription"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type NewSubscriptionUseCase struct {
	authSvc         *auth.Service
	subscriptionSvc *subscription.Service
	logger          *zap.SugaredLogger
}

func NewNewSubscriptionUseCase(authSvc *auth.Service, subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *NewSubscriptionUseCase {
	return &NewSubscriptionUseCase{
		authSvc:         authSvc,
		subscriptionSvc: subscriptionSvc,
		logger:          logger,
	}
}

func (uc *NewSubscriptionUseCase) Exec(ctx context.Context, session string, data *dto.NewSubscriptionReqDTO) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "subscription_usecase.new_subscription")
	defer span.End()

	if err := uc.authSvc.ValidateSession(ctx, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "session validation failed")
		return uuid.Nil, err
	}

	userUUID, err := uc.authSvc.GetSessionData(ctx, session)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve session data")
		return uuid.Nil, err
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
		CreatedAt:     time.Unix(data.CreatedAt, 0),
	}

	subscriptionUUID, err := uc.subscriptionSvc.CreateSubscription(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create subscription")
		return uuid.Nil, err
	}

	span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID.String()))
	span.SetStatus(codes.Ok, "")

	return subscriptionUUID, nil
}
