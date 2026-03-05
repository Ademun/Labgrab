package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("user-service")

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserReq) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "user_service.create_user")
	defer span.End()

	userUUID, err := s.repo.CreateUser(ctx, &DBUser{
		Name:       req.Name,
		Surname:    req.Surname,
		TelegramID: req.TelegramID,
		Username:   req.Username,
		PhotoUrl:   req.PhotoUrl,
	}, req.Tx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create user")
		return uuid.Nil, fmt.Errorf("user service: create user: repository call: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return userUUID, nil
}

func (s *Service) GetUser(ctx context.Context, userUUID uuid.UUID) (*GetUserRes, error) {
	ctx, span := tracer.Start(ctx, "user_service.get_user")
	defer span.End()

	user, err := s.repo.GetUser(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user")
		return nil, fmt.Errorf("user service: get user: repository call: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return &GetUserRes{
		Username:    user.Username,
		Name:        user.Name,
		Surname:     user.Surname,
		Patronymic:  user.Patronymic,
		GroupCode:   user.GroupCode,
		PhoneNumber: user.PhoneNumber,
		TelegramID:  user.TelegramID,
		PhotoUrl:    user.PhotoUrl,
	}, nil
}

func (s *Service) UpdateUser(ctx context.Context, req *UpdateUserReq) error {
	ctx, span := tracer.Start(ctx, "user_service.update_user")
	defer span.End()

	if err := s.repo.UpdateUser(ctx, &DBUser{
		UUID:        req.UserUUID,
		Name:        req.Name,
		Surname:     req.Surname,
		Patronymic:  req.Patronymic,
		GroupCode:   req.GroupCode,
		PhoneNumber: req.PhoneNumber,
		PhotoUrl:    req.PhotoUrl,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update user")
		return fmt.Errorf("user service: update user: repository call: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) ExistsByTelegramID(ctx context.Context, telegramID int) (bool, error) {
	ctx, span := tracer.Start(ctx, "user_service.exists_by_telegram_id")
	defer span.End()

	exists, err := s.repo.ExistsByTelegramID(ctx, telegramID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to check user existence by telegram id")
		return false, fmt.Errorf("user service: exists by telegram id: repository call: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return exists, nil
}

func (s *Service) GetUserUUIDByTelegramID(ctx context.Context, telegramID int) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "user_service.get_user_uuid_by_telegram_id")
	defer span.End()

	userUUID, err := s.repo.GetUserUUIDByTelegramID(ctx, telegramID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user uuid by telegram id")
		return uuid.Nil, fmt.Errorf("user service: get user uuid by telegram id: repository call: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return userUUID, nil
}
