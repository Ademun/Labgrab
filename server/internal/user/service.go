package user

import (
	"context"
	"labgrab/internal/shared/errors"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("user-service")

type Service struct {
	repo   *Repo
	logger *zap.SugaredLogger
}

func NewService(repo *Repo, logger *zap.SugaredLogger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserReq) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "user.service.CreateUser")
	defer span.End()

	user := &DBUser{
		Name:       req.Name,
		Surname:    req.Surname,
		TelegramID: req.TelegramID,
		Username:   req.Username,
		PhotoUrl:   req.PhotoUrl,
	}

	userUUID, err := s.repo.CreateUser(ctx, user, req.Tx)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateUser",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return uuid.Nil, err
	}

	return userUUID, nil
}

func (s *Service) GetUser(ctx context.Context, userUUID uuid.UUID) (*GetUserRes, error) {
	ctx, span := tracer.Start(ctx, "user.service.GetUserInfo")
	defer span.End()

	user, err := s.repo.GetUser(ctx, userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetUserInfo",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
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
	}, nil
}

func (s *Service) UpdateUser(ctx context.Context, req *UpdateUserReq) error {
	ctx, span := tracer.Start(ctx, "user.service.UpdateUserDetails")
	defer span.End()

	user := &DBUser{
		UUID:        req.UserUUID,
		Name:        req.Name,
		Surname:     req.Surname,
		Patronymic:  req.Patronymic,
		GroupCode:   req.GroupCode,
		PhoneNumber: req.PhoneNumber,
		PhotoUrl:    req.PhotoUrl,
	}

	err := s.repo.UpdateUser(ctx, user)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "UpdateUserDetails",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (s *Service) ExistsByTelegramID(ctx context.Context, telegramID int) (bool, error) {
	ctx, span := tracer.Start(ctx, "user.service.ExistsByTelegramID")
	defer span.End()

	exists, err := s.repo.ExistsByTelegramID(ctx, telegramID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "ExistsByTelegramID",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	return exists, nil
}

func (s *Service) GetUserUUIDByTelegramID(ctx context.Context, telegramID int) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "user.service.ExistsByTelegramID")
	defer span.End()

	userUUID, err := s.repo.GetUserUUIDByTelegramID(ctx, telegramID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "ExistsByTelegramID",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return uuid.Nil, err
	}

	return userUUID, nil
}
