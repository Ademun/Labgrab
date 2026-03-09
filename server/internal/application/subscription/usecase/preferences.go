package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/subscription/dto"
	"labgrab/internal/auth"
	"labgrab/internal/subscription"
)

type GetTimePreferencesUsecase struct {
	AuthSvc         *auth.Service
	SubscriptionSvc *subscription.Service
}

func (uc *GetTimePreferencesUsecase) Exec(ctx context.Context, session string) (*dto.GetTimePreferencesResDTO, error) {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("subscription usecase: get time preferences: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("subscription usecase: get time preferences: get session data: %w", err)
	}

	preferences, err := uc.SubscriptionSvc.GetTimePreferences(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription usecase: get time preferences: get time preferences: %w", err)
	}

	return &dto.GetTimePreferencesResDTO{Preferences: preferences}, nil
}

type SetTimePreferencesUsecase struct {
	AuthSvc         *auth.Service
	SubscriptionSvc *subscription.Service
}

func (uc *SetTimePreferencesUsecase) Exec(ctx context.Context, session string, req *dto.SetTimePreferencesReqDTO) error {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return fmt.Errorf("subscription usecase: set time preferences: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return fmt.Errorf("subscription usecase: set time preferences: get session data: %w", err)
	}

	if err := uc.SubscriptionSvc.SetTimePreferences(ctx, userUUID, req.Preferences); err != nil {
		return fmt.Errorf("subscription usecase: set time preferences: set time preferences: %w", err)
	}

	return nil
}

type GetTeacherPreferencesUsecase struct {
	AuthSvc         *auth.Service
	SubscriptionSvc *subscription.Service
}

func (uc *GetTeacherPreferencesUsecase) Exec(ctx context.Context, session string) (*dto.GetTeacherPreferencesResDTO, error) {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("subscription usecase: get teacher preferences: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("subscription usecase: get teacher preferences: get session data: %w", err)
	}

	preferences, err := uc.SubscriptionSvc.GetTeacherPreferences(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription usecase: get teacher preferences: get teacher preferences: %w", err)
	}

	return &dto.GetTeacherPreferencesResDTO{Preferences: preferences}, nil
}

type SetTeacherPreferencesUsecase struct {
	AuthSvc         *auth.Service
	SubscriptionSvc *subscription.Service
}

func (uc *SetTeacherPreferencesUsecase) Exec(ctx context.Context, session string, req *dto.SetTeacherPreferencesReqDTO) error {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return fmt.Errorf("subscription usecase: set teacher preferences: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return fmt.Errorf("subscription usecase: set teacher preferences: get session data: %w", err)
	}

	if err := uc.SubscriptionSvc.SetTeacherPreferences(ctx, userUUID, req.Preferences); err != nil {
		return fmt.Errorf("subscription usecase: set teacher preferences: set teacher preferences: %w", err)
	}

	return nil
}
