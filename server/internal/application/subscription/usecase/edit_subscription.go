package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/auth"
	"labgrab/internal/shared/domain"

	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/subscription"

	"github.com/google/uuid"
)

type EditSubscriptionUsecase struct {
	AuthSvc         *auth.Service
	SubscriptionSvc *subscription.Service
}

func (uc *EditSubscriptionUsecase) Exec(ctx context.Context, session string, subscriptionUUIDStr string, req *dto.EditSubscriptionReqDTO) (string, error) {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return "", fmt.Errorf("subscription usecase: edit subscription: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return "", fmt.Errorf("subscription usecase: edit subscription: get session data: %w", err)
	}

	subscriptionUUID, err := uuid.Parse(subscriptionUUIDStr)
	if err != nil {
		return "", fmt.Errorf("subscription usecase: edit subscription: parse subscription uuid: %w", err)
	}

	existingSub, err := uc.SubscriptionSvc.GetSubscription(ctx, subscriptionUUID)
	if err != nil {
		return "", fmt.Errorf("subscription usecase: edit subscription: get subscription: %w", err)
	}

	updateReq := BuildUpdateReq(userUUID, subscriptionUUID, existingSub, req)

	if err := uc.SubscriptionSvc.UpdateSubscription(ctx, updateReq); err != nil {
		return "", fmt.Errorf("subscription usecase: edit subscription: update subscription: %w", err)
	}

	return subscriptionUUID.String(), nil
}

func BuildUpdateReq(
	userUUID uuid.UUID,
	subscriptionUUID uuid.UUID,
	existing *subscription.GetSubscriptionRes,
	patch *dto.EditSubscriptionReqDTO,
) *subscription.UpdateSubscriptionDataReq {
	req := &subscription.UpdateSubscriptionDataReq{
		UserUUID:         userUUID,
		SubscriptionUUID: subscriptionUUID,
		LabType:          existing.LabType,
		LabTopic:         existing.LabTopic,
		LabNumber:        existing.LabNumber,
		LabAuditorium:    existing.LabAuditorium,
		Status:           existing.Status,
		AutoEnroll:       existing.AutoEnroll,
		AnyDate:          existing.AnyDate,
	}

	if patch.LabType != nil {
		req.LabType = domain.LabType(*patch.LabType)
	}
	if patch.LabTopic != nil {
		req.LabTopic = domain.LabTopic(*patch.LabTopic)
	}
	if patch.LabNumber != nil {
		req.LabNumber = *patch.LabNumber
	}
	if patch.LabAuditorium != nil {
		req.LabAuditorium = patch.LabAuditorium
	}
	if patch.Status != nil {
		req.Status = subscription.Status(*patch.Status)
	}
	if patch.AutoEnroll != nil {
		req.AutoEnroll = *patch.AutoEnroll
	}
	if patch.AnyDate != nil {
		req.AnyDate = *patch.AnyDate
	}

	return req
}
