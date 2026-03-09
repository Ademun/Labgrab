package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/auth"
	"labgrab/internal/user"
)

type UpdateUserUseCase struct {
	AuthSvc *auth.Service
	UserSvc *user.Service
}

func (uc *UpdateUserUseCase) Exec(ctx context.Context, session string, req *dto.UpdateUserReqDTO) error {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return fmt.Errorf("user usecase: update user: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return fmt.Errorf("user usecase: get user: get session data: %w", err)
	}

	data, err := uc.UserSvc.GetUser(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("user usecase: get user: get user: %w", err)
	}

	updateReq := &user.UpdateUserReq{
		UserUUID:    userUUID,
		Name:        data.Name,
		Surname:     data.Surname,
		Patronymic:  data.Patronymic,
		GroupCode:   data.GroupCode,
		PhoneNumber: data.PhoneNumber,
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

	if err := uc.UserSvc.UpdateUser(ctx, updateReq); err != nil {
		return fmt.Errorf("user usecase: update user: update user: %w", err)
	}

	return nil
}
