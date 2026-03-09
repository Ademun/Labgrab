package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/auth"
	"labgrab/internal/subscription"

	"github.com/google/uuid"
)

type GetSubscriptionsUsecase struct {
	AuthSvc         *auth.Service
	SubscriptionSvc *subscription.Service
}

func (uc *GetSubscriptionsUsecase) Exec(ctx context.Context, session string, subscriptionUUIDStr *string) ([]dto.GetSubscriptionsResDTO, error) {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("subscription usecase: get subscriptions: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("subscription usecase: get subscriptions: get session data: %w", err)
	}

	var result []dto.GetSubscriptionsResDTO
	if subscriptionUUIDStr != nil {
		result, err = uc.HandleSingle(ctx, *subscriptionUUIDStr)
	} else {
		result, err = uc.HandleAll(ctx, userUUID)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (uc *GetSubscriptionsUsecase) HandleSingle(ctx context.Context, rawUUID string) ([]dto.GetSubscriptionsResDTO, error) {
	subscriptionUUID, err := uuid.Parse(rawUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription usecase: get subscriptions: parse subscription uuid: %w", err)
	}

	sub, err := uc.SubscriptionSvc.GetSubscription(ctx, subscriptionUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription usecase: get subscriptions: get subscription: %w", err)
	}

	return []dto.GetSubscriptionsResDTO{
		{
			UUID:          sub.SubscriptionUUID.String(),
			LabType:       string(sub.LabType),
			LabTopic:      string(sub.LabTopic),
			LabNumber:     sub.LabNumber,
			LabAuditorium: sub.LabAuditorium,
			Status:        string(sub.Status),
			AutoEnroll:    sub.AutoEnroll,
			AnyDate:       sub.AnyDate,
			CreatedAt:     sub.CreatedAt,
			ClosedAt:      sub.ClosedAt,
		},
	}, nil
}

func (uc *GetSubscriptionsUsecase) HandleAll(ctx context.Context, userUUID uuid.UUID) ([]dto.GetSubscriptionsResDTO, error) {
	subs, err := uc.SubscriptionSvc.GetSubscriptions(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription usecase: get subscriptions: get subscriptions: %w", err)
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
			AutoEnroll:    sub.AutoEnroll,
			AnyDate:       sub.AnyDate,
			CreatedAt:     sub.CreatedAt,
			ClosedAt:      sub.ClosedAt,
		}
	}

	return result, nil
}
