package usecase

import (
	"context"
	"fmt"

	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/subscription"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("subscription-usecase")

type GetSubscriptionsUseCase struct {
	subscriptionSvc *subscription.Service
	logger          *zap.SugaredLogger
}

func NewGetSubscriptionsUseCase(subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *GetSubscriptionsUseCase {
	return &GetSubscriptionsUseCase{
		subscriptionSvc: subscriptionSvc,
		logger:          logger,
	}
}

func (uc *GetSubscriptionsUseCase) Exec(ctx context.Context, data *dto.GetSubscriptionsReqDTO) ([]dto.GetSubscriptionsResDTO, error) {
	ctx, span := tracer.Start(ctx, "subscription.usecase.GetSubscriptions")
	defer span.End()

	userUUID, err := uuid.Parse(data.UserUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid user UUID")
		uc.logger.Errorw("failed to parse user UUID",
			"user_uuid", data.UserUUID,
			"error", err)
		return nil, fmt.Errorf("invalid user UUID: %w", err)
	}

	if data.SubscriptionUUID != nil {
		subscriptionUUID, err := uuid.Parse(*data.SubscriptionUUID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid subscription UUID")
			uc.logger.Errorw("failed to parse subscription UUID",
				"subscription_uuid", *data.SubscriptionUUID,
				"error", err)
			return nil, fmt.Errorf("invalid subscription UUID: %w", err)
		}

		sub, err := uc.subscriptionSvc.GetSubscription(ctx, subscriptionUUID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to get subscription")
			uc.logger.Errorw("failed to get subscription",
				"subscription_uuid", subscriptionUUID,
				"error", err)
			return nil, err
		}

		uc.logger.Infow("retrieved single subscription",
			"user_uuid", userUUID,
			"subscription_uuid", subscriptionUUID)

		return []dto.GetSubscriptionsResDTO{
			{
				UUID:          sub.SubscriptionUUID.String(),
				LabType:       string(sub.LabType),
				LabTopic:      string(sub.LabTopic),
				LabNumber:     sub.LabNumber,
				LabAuditorium: sub.LabAuditorium,
				CreatedAt:     sub.CreatedAt,
				ClosedAt:      sub.ClosedAt,
			},
		}, nil
	}

	subs, err := uc.subscriptionSvc.GetSubscriptions(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get subscriptions")
		uc.logger.Errorw("failed to get subscriptions",
			"user_uuid", userUUID,
			"error", err)
		return nil, err
	}

	uc.logger.Infow("retrieved subscriptions",
		"user_uuid", userUUID,
		"count", len(subs))

	result := make([]dto.GetSubscriptionsResDTO, len(subs))
	for i, sub := range subs {
		result[i] = dto.GetSubscriptionsResDTO{
			UUID:          sub.SubscriptionUUID.String(),
			LabType:       string(sub.LabType),
			LabTopic:      string(sub.LabTopic),
			LabNumber:     sub.LabNumber,
			LabAuditorium: sub.LabAuditorium,
			CreatedAt:     sub.CreatedAt,
			ClosedAt:      sub.ClosedAt,
		}
	}

	return result, nil
}
