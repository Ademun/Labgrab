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

	details := &DBUserDetails{
		Name:       req.Name,
		Surname:    req.Surname,
		Patronymic: req.Patronymic,
		GroupCode:  req.GroupCode,
	}

	contacts := &DBUserContacts{
		PhoneNumber: req.PhoneNumber,
		TelegramID:  req.TelegramID,
	}

	userUUID, err := s.repo.CreateUser(ctx, details, contacts, req.Tx)
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

func (s *Service) GetUserInfo(ctx context.Context, userUUID string) (*GetUserInfoRes, error) {
	ctx, span := tracer.Start(ctx, "user.service.GetUserInfo")
	defer span.End()

	parsedUUID, err := parseUUID(userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetUserInfo",
			Step:      "UUID parsing",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	userInfo, err := s.repo.GetUserInfo(ctx, parsedUUID)
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

	return &GetUserInfoRes{
		UUID:        userInfo.UUID,
		Name:        userInfo.Name,
		Surname:     userInfo.Surname,
		Patronymic:  userInfo.Patronymic,
		GroupCode:   userInfo.GroupCode,
		PhoneNumber: userInfo.PhoneNumber,
		TelegramID:  userInfo.TelegramID,
	}, nil
}

func (s *Service) UpdateUserDetails(ctx context.Context, req *UpdateUserDetailsReq) error {
	ctx, span := tracer.Start(ctx, "user.service.UpdateUserDetails")
	defer span.End()

	details := &DBUserDetails{
		Name:       req.Name,
		Surname:    req.Surname,
		Patronymic: req.Patronymic,
		GroupCode:  req.GroupCode,
		UserUUID:   req.UserUUID,
	}

	err := s.repo.UpdateUserDetails(ctx, details)
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

func (s *Service) UpdateUserContacts(ctx context.Context, req *UpdateUserContactsReq) error {
	ctx, span := tracer.Start(ctx, "user.service.UpdateUserContacts")
	defer span.End()

	contacts := &DBUserContacts{
		PhoneNumber: req.PhoneNumber,
		TelegramID:  req.TelegramID,
		UserUUID:    req.UserUUID,
	}

	err := s.repo.UpdateUserContacts(ctx, contacts)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "UpdateUserContacts",
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

func parseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}
