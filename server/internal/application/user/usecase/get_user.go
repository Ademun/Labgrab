package usecase

import (
	"context"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/auth"
	"labgrab/internal/user"
)

type GetUserUseCase struct {
	authSvc *auth.Service
	userSvc *user.Service
}

func (uc *GetUserUseCase) Exec(ctx context.Context, session string) (*dto.GetUserResDTO, error) {
	if err := uc.authSvc.ValidateSession(ctx, session); err != nil {
		return nil, err
	}

	userUUID, err := uc.authSvc.GetSessionData(ctx, session)
	if err != nil {
		return nil, err
	}

	userData, err := uc.userSvc.GetUser(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return &dto.GetUserResDTO{
		Username:    userData.Username,
		Name:        userData.Name,
		Surname:     userData.Surname,
		Patronymic:  userData.Patronymic,
		GroupCode:   userData.GroupCode,
		PhoneNumber: userData.PhoneNumber,
		PhotoURL:    userData.PhotoUrl,
		ApiReady:    userData.ApiReady,
	}, nil
}
