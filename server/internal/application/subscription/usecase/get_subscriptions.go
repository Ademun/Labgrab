package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/auth"

	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/subscription"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type GetSubscriptionsUseCase struct {
	authSvc         *auth.Service
	subscriptionSvc *subscription.Service
	logger          *zap.SugaredLogger
}

func NewGetSubscriptionsUseCase(authSvc *auth.Service, subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *GetSubscriptionsUseCase {
	return &GetSubscriptionsUseCase{
		authSvc:         authSvc,
		subscriptionSvc: subscriptionSvc,
		logger:          logger,
	}
}

func (uc *GetSubscriptionsUseCase) Exec(ctx context.Context, session string, data *dto.GetSubscriptionsReqDTO) ([]dto.GetSubscriptionsResDTO, error) {
	ctx, span := tracer.Start(ctx, "subscription_usecase.get_subscriptions")
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

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	if data.SubscriptionUUID != nil {
		subscriptionUUID, err := uuid.Parse(*data.SubscriptionUUID)
		if err != nil {
			err = fmt.Errorf("invalid subscription uuid: %w", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid subscription uuid format")
			return nil, err
		}

		span.SetAttributes(
			attribute.String("subscription.uuid", subscriptionUUID.String()),
			attribute.String("query.type", "single"),
		)

		sub, err := uc.subscriptionSvc.GetSubscription(ctx, subscriptionUUID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to retrieve subscription")
			return nil, err
		}

		result := []dto.GetSubscriptionsResDTO{
			{
				UUID:          sub.SubscriptionUUID.String(),
				LabType:       string(sub.LabType),
				LabTopic:      string(sub.LabTopic),
				LabNumber:     sub.LabNumber,
				LabAuditorium: sub.LabAuditorium,
				Status:        string(sub.Status),
				CreatedAt:     sub.CreatedAt,
				ClosedAt:      sub.ClosedAt,
			},
		}

		span.SetStatus(codes.Ok, "")
		return result, nil
	}

	span.SetAttributes(attribute.String("query.type", "all"))

	subs, err := uc.subscriptionSvc.GetSubscriptions(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve subscriptions")
		return nil, err
	}

	result := make([]dto.GetSubscriptionsResDTO, len(subs))
	for i, sub := range subs {
		result[i] = dto.GetSubscriptionsResDTO{
			UUID:          sub.SubscriptionUUID.String(),
			LabType:       string(sub.LabType),
			LabTopic:      string(sub.LabTopic),
			LabNumber:     sub.LabNumber,
			LabAuditorium: sub.LabAuditorium,
			Status:        string(sub.Status),
			CreatedAt:     sub.CreatedAt,
			ClosedAt:      sub.ClosedAt,
		}
	}

	span.SetAttributes(attribute.Int("result.count", len(result)))
	span.SetStatus(codes.Ok, "")

	return result, nil
}
