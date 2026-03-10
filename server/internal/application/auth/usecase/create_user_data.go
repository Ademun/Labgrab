package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/auth/dto"
	"labgrab/internal/auth"
)

type CreateUserDataUsecase struct {
	AuthSvc *auth.Service
}

func (uc *CreateUserDataUsecase) Exec(ctx context.Context, session string, req *dto.CreateUserDataReqDTO) error {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return fmt.Errorf("auth usecase: create user data: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return fmt.Errorf("auth usecase: create user data: get session data: %w", err)
	}

	if err := uc.AuthSvc.CreateUserData(ctx, &auth.CreateUserDataReq{
		UserUUID:          userUUID,
		DikidiPassword:    req.DikidiPassword,
		DikidiPhoneNumber: req.DikidiPhoneNumber,
	}); err != nil {
		return fmt.Errorf("auth usecase: create user data: create user data: %w", err)
	}

	return nil
}
