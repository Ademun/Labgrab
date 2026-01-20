package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

func (s *Service) CreateUser(ctx context.Context, req *CreateUserReq) (*CreateUserRes, error) {
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
		Email:       req.Email,
		TelegramID:  req.TelegramID,
	}

	userUUID, tx, err := s.repo.CreateUser(ctx, details, contacts)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create user")
		s.logger.Errorw("failed to create user in repository", "error", err)
		return nil, fmt.Errorf("%w: %v", ErrCreateUser, err)
	}

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))
	s.logger.Infow("user created successfully", "user_uuid", userUUID)

	return &CreateUserRes{UUID: userUUID, Tx: tx}, nil
}

func (s *Service) GetUserInfo(ctx context.Context, userUUID string) (*GetUserInfoRes, error) {
	ctx, span := tracer.Start(ctx, "user.service.GetUserInfo")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID))

	parsedUUID, err := parseUUID(userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid uuid")
		s.logger.Errorw("invalid user uuid", "uuid", userUUID, "error", err)
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}

	userInfo, err := s.repo.GetUserInfo(ctx, parsedUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			span.SetStatus(codes.Error, "user not found")
			s.logger.Warnw("user not found", "user_uuid", userUUID)
			return nil, ErrUserNotFound
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user info")
		s.logger.Errorw("failed to get user info from repository",
			"user_uuid", userUUID,
			"error", err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	s.logger.Infow("user info retrieved successfully", "user_uuid", userUUID)

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

	span.SetAttributes(attribute.String("user.uuid", req.UserUUID.String()))

	details := &DBUserDetails{
		Name:       req.Name,
		Surname:    req.Surname,
		Patronymic: req.Patronymic,
		GroupCode:  req.GroupCode,
		UserUUID:   req.UserUUID,
	}

	_, validationSpan := tracer.Start(ctx, "user.service.UpdateUserDetails.validate")
	validationErr := s.validateUserDetails(details)
	validationSpan.End()

	if validationErr != nil {
		return s.handleValidationError(validationErr, validationSpan, span, req.UserUUID, "user details")
	}

	err := s.repo.UpdateUserDetails(ctx, details)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update user details")
		s.logger.Errorw("failed to update user details in repository",
			"user_uuid", req.UserUUID,
			"error", err)
		return fmt.Errorf("%w: %v", ErrUpdateUser, err)
	}

	s.logger.Infow("user details updated successfully",
		"user_uuid", req.UserUUID,
		"name", req.Name,
		"surname", req.Surname,
		"group_code", req.GroupCode)

	return nil
}

func (s *Service) UpdateUserContacts(ctx context.Context, req *UpdateUserContactsReq) error {
	ctx, span := tracer.Start(ctx, "user.service.UpdateUserContacts")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", req.UserUUID.String()))

	contacts := &DBUserContacts{
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		TelegramID:  req.TelegramID,
		UserUUID:    req.UserUUID,
	}

	_, validationSpan := tracer.Start(ctx, "user.service.UpdateUserContacts.validate")
	validationErr := s.validateUserContacts(contacts)
	validationSpan.End()

	if validationErr != nil {
		return s.handleValidationError(validationErr, validationSpan, span, req.UserUUID, "user contacts")
	}

	err := s.repo.UpdateUserContacts(ctx, contacts)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update user contacts")
		s.logger.Errorw("failed to update user contacts in repository",
			"user_uuid", req.UserUUID,
			"error", err)
		return fmt.Errorf("%w: %v", ErrUpdateUser, err)
	}

	s.logger.Infow("user contacts updated successfully",
		"user_uuid", req.UserUUID,
		"phone_number", req.PhoneNumber)

	return nil
}

func (s *Service) validateUserDetails(details *DBUserDetails) *ValidationError {
	valErr := NewValidationError()

	if !ValidateAlphabeticString(details.Name) {
		valErr.Add("Name", "must contain only alphabetic characters, spaces, hyphens, underscores, and dots")
	}

	if !ValidateAlphabeticString(details.Surname) {
		valErr.Add("Surname", "must contain only alphabetic characters, spaces, hyphens, underscores, and dots")
	}

	if details.Patronymic != nil && !ValidateAlphabeticString(*details.Patronymic) {
		valErr.Add("Patronymic", "must contain only alphabetic characters, spaces, hyphens, underscores, and dots")
	}

	if !ValidateGroupCode(details.GroupCode) {
		valErr.Add("GroupCode", "must match format XX-YY-ZZ (e.g., AB-12-34)")
	}

	if valErr.HasErrors() {
		return valErr
	}

	return nil
}

func (s *Service) validateUserContacts(contacts *DBUserContacts) *ValidationError {
	valErr := NewValidationError()

	if !ValidatePhoneNumber(contacts.PhoneNumber) {
		valErr.Add("PhoneNumber", "must be in E.164 format (e.g., +1234567890)")
	}

	if contacts.TelegramID != nil && !ValidateTelegramID(int(*contacts.TelegramID)) {
		valErr.Add("TelegramID", "must be a positive integer")
	}

	if valErr.HasErrors() {
		return valErr
	}

	return nil
}

func parseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}

func (s *Service) handleValidationError(validationErr *ValidationError, validationSpan, parentSpan trace.Span, userUUID uuid.UUID, operation string) error {
	validationSpan.RecordError(validationErr)
	validationSpan.SetStatus(codes.Error, "validation failed")
	parentSpan.RecordError(validationErr)
	parentSpan.SetStatus(codes.Error, "validation failed")
	s.logger.Errorw(operation+" validation failed",
		"user_uuid", userUUID,
		"error", validationErr.Error(),
		"validation_errors", validationErr.Errors)
	return validationErr
}
