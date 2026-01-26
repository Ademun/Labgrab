package subscription

import (
	"context"
	"labgrab/internal/shared/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateSubscription",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return subscriptionUUID, err
	}

	return subscriptionUUID, nil
}

func (s *Service) CreateSubscriptionData(ctx context.Context, tx pgx.Tx, req *CreateSubscriptionDataReq) error {
	ctx, span := tracer.Start(ctx, "subscription.service.CreateSubscriptionData")
	defer span.End()

	data := &DBUserSubscriptionData{
		TimePreferences:            req.TimePreferences,
		BlacklistedTeachers:        req.BlacklistedTeachers,
		SuccessfulSubscriptions:    0,
		LastSuccessfulSubscription: nil,
		UserUUID:                   req.UserUUID,
	}

	err := s.repo.CreateSubscriptionData(ctx, tx, data)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateSubscriptionData",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (s *Service) GetSubscription(ctx context.Context, subscriptionUUID uuid.UUID) (*GetSubscriptionRes, error) {
	ctx, span := tracer.Start(ctx, "subscription.service.GetSubscription")
	defer span.End()

	sub, err := s.repo.GetSubscription(ctx, subscriptionUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetSubscription",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

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

	subs, err := s.repo.GetSubscriptions(ctx, userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetSubscriptions",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

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
		err = &errors.ErrServiceProcedure{
			Procedure: "UpdateSubscription",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (s *Service) CloseSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "subscription.service.CloseSubscription")
	defer span.End()

	err := s.repo.CloseSubscription(ctx, subscriptionUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CloseSubscription",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (s *Service) RestoreSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "subscription.service.RestoreSubscription")
	defer span.End()

	err := s.repo.RestoreSubscription(ctx, subscriptionUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "RestoreSubscription",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (s *Service) DeleteSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "subscription.service.DeleteSubscription")
	defer span.End()

	err := s.repo.DeleteSubscription(ctx, subscriptionUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "DeleteSubscription",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (s *Service) GetMatchingSubscriptions(ctx context.Context, req *GetMatchingSubscriptionsReq) ([]GetMatchingSubscriptionsRes, error) {
	ctx, span := tracer.Start(ctx, "subscription.service.GetMatchingSubscriptions")
	defer span.End()

	search := &DBSubscriptionSearch{
		LabType:        req.LabType,
		LabTopic:       req.LabTopic,
		LabNumber:      req.LabNumber,
		LabAuditorium:  req.LabAuditorium,
		AvailableSlots: req.AvailableSlots,
	}

	matches, err := s.repo.GetMatchingSubscriptionsBySlot(ctx, search)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetMatchingSubscriptions",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

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
