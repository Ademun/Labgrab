package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/subscription"
	"time"

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

func (uc *NewSubscriptionUseCase) Exec(ctx context.Context, data *dto.NewSubscriptionReqDTO) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "subscription.usecase.NewSubscription")
	defer span.End()

	userUUID, err := uuid.Parse(data.UserUUID)
	if err != nil {
		err = fmt.Errorf("invalid user uuid: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return uuid.Nil, err
	}

	req := &subscription.CreateSubscriptionReq{
		UserUUID:      userUUID,
		LabType:       subscription.LabType(data.LabType),
		LabTopic:      subscription.LabTopic(data.LabTopic),
		LabNumber:     data.LabNumber,
		LabAuditorium: data.LabAuditorium,
		CreatedAt:     time.Unix(data.CreatedAt, 0),
	}

	subscriptionUUID, err := uc.subscriptionSvc.CreateSubscription(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return uuid.Nil, err
	}

	return subscriptionUUID, nil
}
