package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/auth"
	"labgrab/internal/shared/domain"

	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/subscription"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type EditSubscriptionUseCase struct {
	authSvc         *auth.Service
	subscriptionSvc *subscription.Service
	logger          *zap.SugaredLogger
}

func NewEditSubscriptionUseCase(authSvc *auth.Service, subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *EditSubscriptionUseCase {
	return &EditSubscriptionUseCase{
		authSvc:         authSvc,
		subscriptionSvc: subscriptionSvc,
		logger:          logger,
	}
}

func (uc *EditSubscriptionUseCase) Exec(ctx context.Context, session string, data *dto.EditSubscriptionReqDTO) (*dto.EditSubscriptionResDTO, error) {
	ctx, span := tracer.Start(ctx, "subscription_usecase.edit_subscription")
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

	subscriptionUUID, err := uuid.Parse(data.SubscriptionUUID)
	if err != nil {
		err = fmt.Errorf("invalid subscription uuid: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid subscription uuid format")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.uuid", userUUID.String()),
		attribute.String("subscription.uuid", subscriptionUUID.String()),
	)

	existingSub, err := uc.subscriptionSvc.GetSubscription(ctx, subscriptionUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve existing subscription")
		return nil, err
	}

	labType := existingSub.LabType
	if data.LabType != nil {
		labType = domain.LabType(*data.LabType)
	}

	labTopic := existingSub.LabTopic
	if data.LabTopic != nil {
		labTopic = domain.LabTopic(*data.LabTopic)
	}

	labNumber := existingSub.LabNumber
	if data.LabNumber != nil {
		labNumber = *data.LabNumber
	}

	labAuditorium := existingSub.LabAuditorium
	if data.LabAuditorium != nil {
		labAuditorium = data.LabAuditorium
	}

	status := existingSub.Status
	if data.Status != nil {
		status = subscription.Status(*data.Status)
	}

	autoEnroll := existingSub.AutoEnroll
	if data.AutoEnroll != nil {
		autoEnroll = *data.AutoEnroll
	}

	anyDate := existingSub.AnyDate
	if data.AnyDate != nil {
		anyDate = *data.AnyDate
	}

	req := &subscription.UpdateSubscriptionDataReq{
		UserUUID:         userUUID,
		SubscriptionUUID: subscriptionUUID,
		LabType:          labType,
		LabTopic:         labTopic,
		LabNumber:        labNumber,
		LabAuditorium:    labAuditorium,
		Status:           status,
		AutoEnroll:       autoEnroll,
		AnyDate:          anyDate,
	}

	if err := uc.subscriptionSvc.UpdateSubscription(ctx, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update subscription")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")

	return &dto.EditSubscriptionResDTO{
		UUID: subscriptionUUID.String(),
	}, nil
}
