package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/auth/dto"
	"labgrab/internal/auth"
	"labgrab/internal/user"
)

type AuthUserUsecase struct {
	AuthSvc *auth.Service
	UserSvc *user.Service
}

func (uc *AuthUserUsecase) Exec(ctx context.Context, req *dto.AuthUserReqDTO) (string, error) {
	if err := uc.AuthSvc.ValidateTelegramAuthData(ctx, &auth.ValidateTelegramAuthDataReq{
		Id:        req.Id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		PhotoURL:  req.PhotoURL,
		AuthDate:  req.AuthDate,
		Hash:      req.Hash,
	}); err != nil {
		return "", fmt.Errorf("auth usecase: auth user: validate auth data: %w", err)
	}

	exists, err := uc.UserSvc.ExistsByTelegramID(ctx, req.Id)
	if err != nil {
		return "", fmt.Errorf("auth usecase: auth user: check if user exists: %w", err)
	}

	if exists {
		return uc.HandleExistingUser(ctx, req)
	}

	return uc.HandleNewUser(ctx, req)
}

func (uc *AuthUserUsecase) HandleNewUser(ctx context.Context, req *dto.AuthUserReqDTO) (string, error) {
	userUUID, err := uc.UserSvc.CreateUser(ctx, &user.CreateUserReq{
		TelegramID: req.Id,
		Name:       emptishStringToPtr(req.FirstName),
		Surname:    emptishStringToPtr(req.LastName),
		Username:   req.Username,
		PhotoUrl:   nil,
	})
	if err != nil {
		return "", fmt.Errorf("auth usecase: auth user: create user: %w", err)
	}

	session, err := uc.AuthSvc.CreateSession(ctx, userUUID)
	if err != nil {
		return "", fmt.Errorf("auth usecase: auth user: create session: %w", err)
	}

	return session, nil
}

func (uc *AuthUserUsecase) HandleExistingUser(ctx context.Context, req *dto.AuthUserReqDTO) (string, error) {
	userUUID, err := uc.UserSvc.GetUserUUIDByTelegramID(ctx, req.Id)
	if err != nil {
		return "", fmt.Errorf("auth usecase: auth user: get user uuid by telegram id: %w", err)
	}

	session, err := uc.AuthSvc.CreateSession(ctx, userUUID)
	if err != nil {
		return "", fmt.Errorf("auth usecase: auth user: create session: %w", err)
	}

	return session, nil
}

func emptishStringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
