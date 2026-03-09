package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/auth"
)

type AuthStaleUsersUsecase struct {
	AuthSvc *auth.Service
}

func (uc *AuthStaleUsersUsecase) Exec(ctx context.Context) error {
	if err := uc.AuthSvc.AuthStaleUsers(ctx); err != nil {
		return fmt.Errorf("auth stale users usecase: %w", err)
	}

	return nil
}
