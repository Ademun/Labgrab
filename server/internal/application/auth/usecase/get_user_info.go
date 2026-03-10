package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/auth/dto"
	"labgrab/internal/auth"
)

type GetUserInfoUsecase struct {
	AuthSvc *auth.Service
}

func (uc *GetUserInfoUsecase) Exec(ctx context.Context, session string) (*dto.GetUserInfoResDTO, error) {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("auth usecase: get user info: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("auth usecase: get user info: get session data: %w", err)
	}

	info, err := uc.AuthSvc.GetUserInfo(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("auth usecase: get user info: get user info: %w", err)
	}

	if info == nil {
		return nil, nil
	}

	return &dto.GetUserInfoResDTO{
		PhoneNumber: info.DikidiPhoneNumber,
		Password:    info.DikidiPassword,
		ApiAuthed:   info.ApiAuthed,
		LastAuth:    info.LastAuth,
	}, nil
}
