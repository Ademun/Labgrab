package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/auth/dto"
	"labgrab/internal/auth"
)

type UpdateUserDataUsecase struct {
	AuthSvc *auth.Service
}

func (uc *UpdateUserDataUsecase) Exec(ctx context.Context, session string, req *dto.UpdateUserDataReqDTO) error {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return fmt.Errorf("auth usecase: update user data: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return fmt.Errorf("auth usecase: update user data: get session data: %w", err)
	}

	if err := uc.AuthSvc.UpdateUserData(ctx, &auth.UpdateUserDataReq{
		UserUUID:          userUUID,
		DikidiPassword:    req.DikidiPassword,
		DikidiPhoneNumber: req.DikidiPhoneNumber,
	}); err != nil {
		return fmt.Errorf("auth usecase: update user data: create user data: %w", err)
	}

	return nil
}
