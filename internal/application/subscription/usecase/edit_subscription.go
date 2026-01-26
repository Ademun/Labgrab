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

type EditSubscriptionUseCase struct {
	subscriptionSvc *subscription.Service
	logger          *zap.SugaredLogger
}

func NewEditSubscriptionUseCase(subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *EditSubscriptionUseCase {
	return &EditSubscriptionUseCase{
		subscriptionSvc: subscriptionSvc,
		logger:          logger,
	}
}

func (uc *EditSubscriptionUseCase) Exec(ctx context.Context, data *dto.EditSubscriptionReqDTO) (*dto.EditSubscriptionResDTO, error) {
	ctx, span := tracer.Start(ctx, "subscription.usecase.EditSubscription")
	defer span.End()

	userUUID, err := uuid.Parse(data.UserUUID)
	if err != nil {
		err = fmt.Errorf("invalid user uuid: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	subscriptionUUID, err := uuid.Parse(data.SubscriptionUUID)
	if err != nil {
		err = fmt.Errorf("invalid subscription uuid: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	existingSub, err := uc.subscriptionSvc.GetSubscription(ctx, subscriptionUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	labType := existingSub.LabType
	if data.LabType != nil {
		labType = subscription.LabType(*data.LabType)
	}

	labTopic := existingSub.LabTopic
	if data.LabTopic != nil {
		labTopic = subscription.LabTopic(*data.LabTopic)
	}

	labNumber := existingSub.LabNumber
	if data.LabNumber != nil {
		labNumber = *data.LabNumber
	}

	labAuditorium := existingSub.LabAuditorium
	if data.LabAuditorium != nil {
		labAuditorium = data.LabAuditorium
	}

	req := &subscription.UpdateSubscriptionDataReq{
		UserUUID:         userUUID,
		SubscriptionUUID: subscriptionUUID,
		LabType:          labType,
		LabTopic:         labTopic,
		LabNumber:        labNumber,
		LabAuditorium:    labAuditorium,
	}

	if err := uc.subscriptionSvc.UpdateSubscription(ctx, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &dto.EditSubscriptionResDTO{
		UUID: subscriptionUUID.String(),
	}, nil
}
