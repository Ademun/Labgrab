package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/auth"
)

type DikidiAuthUsecase struct {
	AuthSvc *auth.Service
}

func (uc *DikidiAuthUsecase) Exec(ctx context.Context, session string) error {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return fmt.Errorf("auth usecase: dikidi auth: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return fmt.Errorf("auth usecase: dikidi auth: get session data: %w", err)
	}

	if err := uc.AuthSvc.AuthUser(ctx, userUUID); err != nil {
		return fmt.Errorf("auth usecase: dikidi auth: auth user: %w", err)
	}

	return nil
}
