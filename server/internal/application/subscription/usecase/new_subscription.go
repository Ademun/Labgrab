package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/auth"
	"labgrab/internal/shared/domain"
	"labgrab/internal/subscription"
	"time"
)

type NewSubscriptionUsecase struct {
	AuthSvc         *auth.Service
	SubscriptionSvc *subscription.Service
}

func (uc *NewSubscriptionUsecase) Exec(ctx context.Context, session string, req *dto.NewSubscriptionReqDTO) (string, error) {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return "", fmt.Errorf("subscription usecase: new subscription: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return "", fmt.Errorf("subscription usecase: new subscription: get session data: %w", err)
	}

	subscriptionUUID, err := uc.SubscriptionSvc.CreateSubscription(ctx, &subscription.CreateSubscriptionReq{
		UserUUID:      userUUID,
		LabType:       domain.LabType(req.LabType),
		LabTopic:      domain.LabTopic(req.LabTopic),
		LabNumber:     req.LabNumber,
		LabAuditorium: req.LabAuditorium,
		AutoEnroll:    req.AutoEnroll,
		AnyDate:       req.AnyDate,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		return "", fmt.Errorf("subscription usecase: new subscription: create subscription: %w", err)
	}

	return subscriptionUUID.String(), nil
}
