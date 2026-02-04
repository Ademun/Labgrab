package usecase

import (
	"context"
	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/auth"
	"labgrab/internal/subscription"
	"time"

	"github.com/google/uuid"
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
	if err := uc.authSvc.ValidateSession(ctx, session); err != nil {
		return uuid.Nil, err
	}

	userUUID, err := uc.authSvc.GetSessionData(ctx, session)
	if err != nil {
		return uuid.Nil, err
	}

	ctx, span := tracer.Start(ctx, "subscription.usecase.NewSubscription")
	defer span.End()

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
