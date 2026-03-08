package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserReq) (uuid.UUID, error) {
	userUUID, err := s.repo.CreateUser(ctx, &DBUser{
		Name:       req.Name,
		Surname:    req.Surname,
		TelegramID: req.TelegramID,
		Username:   req.Username,
		PhotoUrl:   req.PhotoUrl,
	}, req.Tx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user service: create user: repository call: %w", err)
	}
	return userUUID, nil
}

func (s *Service) GetUser(ctx context.Context, userUUID uuid.UUID) (*GetUserRes, error) {
	user, err := s.repo.GetUser(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("user service: get user: repository call: %w", err)
	}
	return &GetUserRes{
		Username:    user.Username,
		Name:        user.Name,
		Surname:     user.Surname,
		Patronymic:  user.Patronymic,
		GroupCode:   user.GroupCode,
		PhoneNumber: user.PhoneNumber,
		TelegramID:  user.TelegramID,
		PhotoUrl:    user.PhotoUrl,
		ApiReady:    user.ApiReady,
	}, nil
}

func (s *Service) UpdateUser(ctx context.Context, req *UpdateUserReq) error {
	if err := s.repo.UpdateUser(ctx, &DBUser{
		UUID:        req.UserUUID,
		Name:        req.Name,
		Surname:     req.Surname,
		Patronymic:  req.Patronymic,
		GroupCode:   req.GroupCode,
		PhoneNumber: req.PhoneNumber,
		PhotoUrl:    req.PhotoUrl,
	}); err != nil {
		return fmt.Errorf("user service: update user: repository call: %w", err)
	}
	return nil
}

func (s *Service) ExistsByTelegramID(ctx context.Context, telegramID int) (bool, error) {
	exists, err := s.repo.ExistsByTelegramID(ctx, telegramID)
	if err != nil {
		return false, fmt.Errorf("user service: exists by telegram id: repository call: %w", err)
	}
	return exists, nil
}

func (s *Service) GetUserUUIDByTelegramID(ctx context.Context, telegramID int) (uuid.UUID, error) {
	userUUID, err := s.repo.GetUserUUIDByTelegramID(ctx, telegramID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user service: get user uuid by telegram id: repository call: %w", err)
	}
	return userUUID, nil
}
