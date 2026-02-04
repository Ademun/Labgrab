package usecase

import (
	"context"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/auth"
	"labgrab/internal/user"
)

type UpdateUserUseCase struct {
	authSvc *auth.Service
	userSvc *user.Service
}

func NewUpdateUserUseCase(authSvc *auth.Service, userSvc *user.Service) *UpdateUserUseCase {
	return &UpdateUserUseCase{authSvc: authSvc, userSvc: userSvc}
}

func (uc *UpdateUserUseCase) Exec(ctx context.Context, session string, req *dto.UpdateUserReqDTO) error {
	if err := uc.authSvc.ValidateSession(ctx, session); err != nil {
		return err
	}

	userUUID, err := uc.authSvc.GetSessionData(ctx, session)
	if err != nil {
		return err
	}

	data, err := uc.userSvc.GetUser(ctx, userUUID)
	if err != nil {
		return err
	}

	updateReq := &user.UpdateUserReq{
		UserUUID:    userUUID,
		Name:        data.Name,
		Surname:     data.Surname,
		Patronymic:  data.Patronymic,
		GroupCode:   data.GroupCode,
		PhoneNumber: data.PhoneNumber,
		PhotoUrl:    data.PhotoUrl,
	}

	if req.Name != nil {
		updateReq.Name = req.Name
	}

	if req.Surname != nil {
		updateReq.Surname = req.Surname
	}

	if req.Patronymic != nil {
		updateReq.Patronymic = req.Patronymic
	}

	if req.GroupCode != nil {
		updateReq.GroupCode = req.GroupCode
	}

	if req.PhoneNumber != nil {
		updateReq.PhoneNumber = req.PhoneNumber
	}

	if req.PhotoUrl != nil {
		updateReq.PhotoUrl = req.PhotoUrl
	}

	if err := uc.userSvc.UpdateUser(ctx, updateReq); err != nil {
		return err
	}

	return nil
}
