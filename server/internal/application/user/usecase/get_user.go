package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/auth"
	"labgrab/internal/user"
)

type GetUserUseCase struct {
	AuthSvc *auth.Service
	UserSvc *user.Service
}

func (uc *GetUserUseCase) Exec(ctx context.Context, session string) (*dto.GetUserResDTO, error) {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("user usecase: get user: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("user usecase: get user: get session data: %w", err)
	}

	userData, err := uc.UserSvc.GetUser(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("user usecase: get user: get user: %w", err)
	}

	return &dto.GetUserResDTO{
		Name:             userData.Name,
		Surname:          userData.Surname,
		Patronymic:       userData.Patronymic,
		GroupCode:        userData.GroupCode,
		PhoneNumber:      userData.PhoneNumber,
		TelegramPhotoURL: userData.TelegramPhotoUrl,
		TelegramUsername: userData.TelegramUsername,
		ApiReady:         userData.ApiReady,
	}, nil
}
