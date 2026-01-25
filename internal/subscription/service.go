package subscription

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

var tracer = otel.Tracer("subscription-service")

type Service struct {
	repo   *Repo
	logger *zap.SugaredLogger
}

func NewService(repo *Repo, logger *zap.SugaredLogger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) CreateSubscription(ctx context.Context, req *CreateSubscriptionReq) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "subscription.service.CreateSubscription")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", req.UserUUID.String()))

	_, validationSpan := tracer.Start(ctx, "subscription.service.CreateSubscription.validate")
	validationErr := s.validateSubscriptionReq(req.LabType, req.LabTopic, req.LabNumber, req.LabAuditorium)
	validationSpan.End()

	if validationErr != nil {
		return uuid.Nil, s.handleValidationError(validationErr, validationSpan, span, req.UserUUID, "subscription")
	}

	dbSub := &DBSubscription{
		LabType:       req.LabType,
		LabTopic:      req.LabTopic,
		LabNumber:     req.LabNumber,
		LabAuditorium: req.LabAuditorium,
		CreatedAt:     req.CreatedAt,
		ClosedAt:      nil,
		UserUUID:      req.UserUUID,
	}

	subscriptionUUID, err := s.repo.CreateSubscription(ctx, dbSub)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create subscription")
		s.logger.Errorw("failed to create subscription in repository",
			"user_uuid", req.UserUUID,
			"subscription_uuid", subscriptionUUID,
			"error", err)
		return subscriptionUUID, fmt.Errorf("%w: %v", ErrCreateSubscription, err)
	}

	span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID.String()))
	s.logger.Infow("subscription created successfully",
		"user_uuid", req.UserUUID,
		"subscription_uuid", subscriptionUUID,
		"lab_type", req.LabType,
		"lab_topic", req.LabTopic,
		"lab_number", req.LabNumber)

	return subscriptionUUID, nil
}

func (s *Service) CreateSubscriptionData(ctx context.Context, tx pgx.Tx, req *CreateSubscriptionDataReq) error {
	ctx, span := tracer.Start(ctx, "subscription.service.CreateSubscriptionData")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", req.UserUUID.String()))

	data := &DBUserSubscriptionData{
		TimePreferences:            req.TimePreferences,
		BlacklistedTeachers:        req.BlacklistedTeachers,
		SuccessfulSubscriptions:    0,
		LastSuccessfulSubscription: nil,
		UserUUID:                   req.UserUUID,
	}

	err := s.repo.CreateSubscriptionData(ctx, tx, data)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create subscription data")
		s.logger.Errorw("failed to create subscription data in repository",
			"user_uuid", req.UserUUID,
			"error", err)
		return fmt.Errorf("%w: %v", ErrCreateSubscription, err)
	}

	s.logger.Infow("subscription data created successfully",
		"user_uuid", req.UserUUID)

	return nil
}

func (s *Service) GetSubscription(ctx context.Context, subscriptionUUID uuid.UUID) (*GetSubscriptionRes, error) {
	ctx, span := tracer.Start(ctx, "subscription.service.GetSubscription")
	defer span.End()

	span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID.String()))

	sub, err := s.repo.GetSubscription(ctx, subscriptionUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			span.SetStatus(codes.Error, "subscription not found")
			s.logger.Warnw("subscription not found", "subscription_uuid", subscriptionUUID)
			return nil, ErrSubscriptionNotFound
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get subscription")
		s.logger.Errorw("failed to get subscription from repository",
			"subscription_uuid", subscriptionUUID,
			"error", err)
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	s.logger.Infow("subscription retrieved successfully", "subscription_uuid", subscriptionUUID)

	return &GetSubscriptionRes{
		SubscriptionUUID: sub.SubscriptionUUID,
		LabType:          sub.LabType,
		LabTopic:         sub.LabTopic,
		LabNumber:        sub.LabNumber,
		LabAuditorium:    sub.LabAuditorium,
		CreatedAt:        sub.CreatedAt,
		ClosedAt:         sub.ClosedAt,
	}, nil
}

func (s *Service) GetSubscriptions(ctx context.Context, userUUID uuid.UUID) ([]GetSubscriptionRes, error) {
	ctx, span := tracer.Start(ctx, "subscription.service.GetSubscriptions")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	subs, err := s.repo.GetSubscriptions(ctx, userUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get subscriptions")
		s.logger.Errorw("failed to get subscriptions from repository",
			"user_uuid", userUUID,
			"error", err)
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	s.logger.Infow("subscriptions retrieved successfully",
		"user_uuid", userUUID,
		"count", len(subs))

	result := make([]GetSubscriptionRes, len(subs))
	for i, sub := range subs {
		result[i] = GetSubscriptionRes{
			SubscriptionUUID: sub.SubscriptionUUID,
			LabType:          sub.LabType,
			LabTopic:         sub.LabTopic,
			LabNumber:        sub.LabNumber,
			LabAuditorium:    sub.LabAuditorium,
			CreatedAt:        sub.CreatedAt,
			ClosedAt:         sub.ClosedAt,
		}
	}

	return result, nil
}

func (s *Service) UpdateSubscription(ctx context.Context, req *UpdateSubscriptionDataReq) error {
	ctx, span := tracer.Start(ctx, "subscription.service.UpdateSubscription")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", req.UserUUID.String()))

	_, validationSpan := tracer.Start(ctx, "subscription.service.UpdateSubscription.validate")
	validationErr := s.validateSubscriptionReq(req.LabType, req.LabTopic, req.LabNumber, req.LabAuditorium)
	validationSpan.End()

	if validationErr != nil {
		return s.handleValidationError(validationErr, validationSpan, span, req.UserUUID, "subscription")
	}

	subscription := &DBSubscription{
		SubscriptionUUID: req.SubscriptionUUID,
		LabType:          req.LabType,
		LabTopic:         req.LabTopic,
		LabNumber:        req.LabNumber,
		LabAuditorium:    req.LabAuditorium,
		UserUUID:         req.UserUUID,
	}

	err := s.repo.UpdateSubscription(ctx, subscription)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update subscription")
		s.logger.Errorw("failed to update subscription in repository",
			"user_uuid", req.UserUUID,
			"subscription_uuid", req.SubscriptionUUID,
			"error", err)
		return fmt.Errorf("%w: %v", ErrUpdateSubscription, err)
	}

	s.logger.Infow("subscription updated successfully",
		"user_uuid", req.UserUUID,
		"subscription_uuid", req.SubscriptionUUID,
		"lab_type", req.LabType,
		"lab_topic", req.LabTopic,
		"lab_number", req.LabNumber)

	return nil
}

func (s *Service) CloseSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "subscription.service.CloseSubscription")
	defer span.End()

	span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID.String()))

	err := s.repo.CloseSubscription(ctx, subscriptionUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to close subscription")
		s.logger.Errorw("failed to close subscription in repository",
			"subscription_uuid", subscriptionUUID,
			"error", err)
		return fmt.Errorf("%w: %v", ErrCloseSubscription, err)
	}

	s.logger.Infow("subscription closed successfully",
		"subscription_uuid", subscriptionUUID)

	return nil
}

func (s *Service) RestoreSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "subscription.service.RestoreSubscription")
	defer span.End()

	span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID.String()))

	err := s.repo.RestoreSubscription(ctx, subscriptionUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to restore subscription")
		s.logger.Errorw("failed to restore subscription in repository",
			"subscription_uuid", subscriptionUUID,
			"error", err)
		return fmt.Errorf("%w: %v", ErrRestoreSubscription, err)
	}

	s.logger.Infow("subscription restored successfully",
		"subscription_uuid", subscriptionUUID)

	return nil
}

func (s *Service) DeleteSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "subscription.service.DeleteSubscription")
	defer span.End()

	span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID.String()))

	err := s.repo.DeleteSubscription(ctx, subscriptionUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete subscription")
		s.logger.Errorw("failed to delete subscription in repository",
			"subscription_uuid", subscriptionUUID,
			"error", err)
		return fmt.Errorf("%w: %v", ErrDeleteSubscription, err)
	}

	s.logger.Infow("subscription deleted successfully",
		"subscription_uuid", subscriptionUUID)

	return nil
}

func (s *Service) GetMatchingSubscriptions(ctx context.Context, req *GetMatchingSubscriptionsReq) ([]GetMatchingSubscriptionsRes, error) {
	ctx, span := tracer.Start(ctx, "subscription.service.GetMatchingSubscriptions")
	defer span.End()

	span.SetAttributes(
		attribute.String("lab.type", string(req.LabType)),
		attribute.String("lab.topic", string(req.LabTopic)),
		attribute.Int("lab.number", req.LabNumber),
		attribute.Int("lab.auditorium", req.LabAuditorium),
	)

	search := &DBSubscriptionSearch{
		LabType:        req.LabType,
		LabTopic:       req.LabTopic,
		LabNumber:      req.LabNumber,
		LabAuditorium:  req.LabAuditorium,
		AvailableSlots: req.AvailableSlots,
	}

	matches, err := s.repo.GetMatchingSubscriptionsBySlot(ctx, search)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get matching subscriptions")
		s.logger.Errorw("failed to get matching subscriptions from repository",
			"lab_type", req.LabType,
			"lab_topic", req.LabTopic,
			"lab_number", req.LabNumber,
			"error", err)
		return nil, fmt.Errorf("failed to get matching subscriptions: %w", err)
	}

	s.logger.Infow("matching subscriptions retrieved successfully",
		"lab_type", req.LabType,
		"lab_topic", req.LabTopic,
		"lab_number", req.LabNumber,
		"matches_count", len(matches))

	result := make([]GetMatchingSubscriptionsRes, len(matches))
	for i, match := range matches {
		result[i] = GetMatchingSubscriptionsRes{
			UserUUID:                   match.UserUUID,
			SubscriptionUUID:           match.SubscriptionUUID,
			SuccessfulSubscriptions:    match.SuccessfulSubscriptions,
			LastSuccessfulSubscription: match.LastSuccessfulSubscription,
			MatchingTimeslots:          match.MatchingTimeslots,
		}
	}

	return result, nil
}

func (s *Service) validateSubscriptionReq(labType LabType, labTopic LabTopic, labNumber int, labAuditorium *int) *ValidationError {
	valErr := NewValidationError()

	if !ValidateLabType(labType) {
		valErr.Add("LabType", "must be either Defence or Performance")
	}

	if !ValidateLabTopic(labTopic) {
		valErr.Add("LabTopic", "must be one of Virtual, Electricity, or Mechanics")
	}

	if !ValidateLabNumber(labNumber) {
		valErr.Add("LabNumber", "must be between 1 and 255")
	}

	if labType == LabTypePerformance && labAuditorium == nil {
		valErr.Add("LabAuditorium", "must not be nil for Performance lab type")
	}

	if labType == LabTypeDefence && labAuditorium != nil {
		valErr.Add("LabAuditorium", "must be nil for Defence lab type")
	}

	if valErr.HasErrors() {
		return valErr
	}

	return nil
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
