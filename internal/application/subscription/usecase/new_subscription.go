package usecase

import (
	"context"
	"fmt"

	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/subscription"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type NewSubscriptionUseCase struct {
	subscriptionSvc *subscription.Service
	logger          *zap.SugaredLogger
}

func NewNewSubscriptionUseCase(subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *NewSubscriptionUseCase {
	return &NewSubscriptionUseCase{
		subscriptionSvc: subscriptionSvc,
		logger:          logger,
	}
}

func (uc *NewSubscriptionUseCase) Exec(ctx context.Context, data *dto.NewSubscriptionReqDTO) (*dto.NewSubscriptionResDTO, error) {
	ctx, span := tracer.Start(ctx, "subscription.usecase.NewSubscription")
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

	req := &subscription.CreateSubscriptionReq{
		UserUUID:      userUUID,
		LabType:       subscription.LabType(data.LabType),
		LabTopic:      subscription.LabTopic(data.LabTopic),
		LabNumber:     data.LabNumber,
		LabAuditorium: data.LabAuditorium,
		CreatedAt:     data.CreatedAt,
	}

	if err := uc.subscriptionSvc.CreateSubscription(ctx, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create subscription")
		uc.logger.Errorw("failed to create subscription",
			"user_uuid", userUUID,
			"error", err)
		return nil, err
	}

	uc.logger.Infow("subscription created successfully",
		"user_uuid", userUUID)

	return &dto.NewSubscriptionResDTO{
		UUID: userUUID.String(),
	}, nil
}
