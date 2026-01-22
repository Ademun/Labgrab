package usecase

import (
	"context"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/auth"
	"labgrab/internal/user"
)

type AuthUserUseCase struct {
	authSvc *auth.Service
	userSvc *user.Service
}

func NewAuthUserUseCase(authSvc *auth.Service, userSvc *user.Service) *AuthUserUseCase {
	return &AuthUserUseCase{authSvc: authSvc, userSvc: userSvc}
}

func (uc *AuthUserUseCase) Exec(ctx context.Context, data *dto.AuthUserReqDTO) (*dto.AuthUserRespDTO, error) {
	// telegramAuthData := &auth.TelegramAuthData{
	// 	Id:        data.Id,
	// 	FirstName: data.FirstName,
	// 	LastName:  data.LastName,
	// 	Username:  data.Username,
	// 	PhotoURL:  data.PhotoURL,
	// 	AuthDate:  data.AuthDate,
	// 	Hash:      data.Hash,
	// }
	//
	// if err := uc.authSvc.ValidateTelegramAuthData(ctx, telegramAuthData); err != nil {
	//
	// }

	exists, err := uc.userSvc.ExistsByTelegramID(ctx, data.Id)
	if err != nil {
		return nil, err
	}

	return &dto.AuthUserRespDTO{Exists: exists}, nil
}
