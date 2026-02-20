package subscription

import (
	"context"
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/errors"
	"labgrab/internal/shared/types"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("subscription-service")

type Service struct {
	repo         *Repo
	deduplicator *Deduplicator
	logger       *zap.SugaredLogger
}

func NewService(repo *Repo, deduplicator *Deduplicator, logger *zap.SugaredLogger) *Service {
	return &Service{repo: repo, deduplicator: deduplicator, logger: logger}
}

func (s *Service) CreateSubscription(ctx context.Context, req *CreateSubscriptionReq) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "subscription_service.create_subscription")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUUID.String()),
		attribute.String("lab.type", string(req.LabType)),
		attribute.String("lab.topic", string(req.LabTopic)),
		attribute.Int("lab.number", req.LabNumber),
	)

	if req.LabAuditorium != nil {
		span.SetAttributes(attribute.Int("lab.auditorium", *req.LabAuditorium))
	}

	if err := req.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return uuid.Nil, err
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
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateSubscription",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create subscription in repository")
		return uuid.Nil, err
	}

	span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID.String()))
	span.SetStatus(codes.Ok, "")
	return subscriptionUUID, nil
}

func (s *Service) GetSubscription(ctx context.Context, subscriptionUUID uuid.UUID) (*GetSubscriptionRes, error) {
	ctx, span := tracer.Start(ctx, "subscription_service.get_subscription")
	defer span.End()

	span.SetAttributes(attribute.String("subscription.uuid", subscriptionUUID.String()))

	sub, err := s.repo.GetSubscription(ctx, subscriptionUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetSubscription",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve subscription from repository")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.uuid", sub.UserUUID.String()),
		attribute.String("lab.type", string(sub.LabType)),
		attribute.String("lab.topic", string(sub.LabTopic)),
		attribute.Int("lab.number", sub.LabNumber),
	)

	span.SetStatus(codes.Ok, "")
	return &GetSubscriptionRes{
		SubscriptionUUID: sub.SubscriptionUUID,
		LabType:          sub.LabType,
		LabTopic:         sub.LabTopic,
		LabNumber:        sub.LabNumber,
		LabAuditorium:    sub.LabAuditorium,
		Status:           sub.Status,
		AutoEnroll:       sub.AutoEnroll,
		AnyDate:          sub.AnyDate,
		CreatedAt:        sub.CreatedAt,
		ClosedAt:         sub.ClosedAt,
	}, nil
}

func (s *Service) GetSubscriptions(ctx context.Context, userUUID uuid.UUID) ([]GetSubscriptionRes, error) {
	ctx, span := tracer.Start(ctx, "subscription_service.get_subscriptions")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	subs, err := s.repo.GetSubscriptions(ctx, userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetSubscriptions",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve subscriptions from repository")
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
			Status:           sub.Status,
			AutoEnroll:       sub.AutoEnroll,
			AnyDate:          sub.AnyDate,
			CreatedAt:        sub.CreatedAt,
			ClosedAt:         sub.ClosedAt,
		}
	}

	span.SetAttributes(attribute.Int("subscriptions.count", len(result)))
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (s *Service) UpdateSubscription(ctx context.Context, req *UpdateSubscriptionDataReq) error {
	ctx, span := tracer.Start(ctx, "subscription_service.update_subscription")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUUID.String()),
		attribute.String("subscription.uuid", req.SubscriptionUUID.String()),
		attribute.String("lab.type", string(req.LabType)),
		attribute.String("lab.topic", string(req.LabTopic)),
		attribute.Int("lab.number", req.LabNumber),
	)

	if req.LabAuditorium != nil {
		span.SetAttributes(attribute.Int("lab.auditorium", *req.LabAuditorium))
	}

	if err := req.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return err
	}

	subscription := &DBSubscription{
		SubscriptionUUID: req.SubscriptionUUID,
		LabType:          req.LabType,
		LabTopic:         req.LabTopic,
		LabNumber:        req.LabNumber,
		LabAuditorium:    req.LabAuditorium,
		Status:           req.Status,
		AutoEnroll:       req.AutoEnroll,
		AnyDate:          req.AnyDate,
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
		span.SetStatus(codes.Error, "failed to update subscription in repository")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) GetTimePreferences(ctx context.Context, userUUID uuid.UUID) (UserTimePreferences, error) {
	ctx, span := tracer.Start(ctx, "subscription_service.get_time_preferences")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	dbPrefs, err := s.repo.GetTimePreferences(ctx, userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetTimePreferences",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve time preferences from repository")
		return nil, err
	}

	// Transform []DBTimePreferences to map[int]map[types.DayOfWeek][]types.Lesson
	userPrefs := make(UserTimePreferences)
	for _, pref := range dbPrefs {
		if _, exists := userPrefs[pref.WeekNumber]; !exists {
			userPrefs[pref.WeekNumber] = make(map[types.DayOfWeek][]domain.Lesson)
		}
		userPrefs[pref.WeekNumber][pref.DayOfWeek] = pref.Lessons
	}

	span.SetAttributes(
		attribute.Int("preferences.weeks_count", len(userPrefs)),
		attribute.Int("preferences.records_count", len(dbPrefs)),
	)
	span.SetStatus(codes.Ok, "")

	return userPrefs, nil
}

func (s *Service) SetTimePreferences(ctx context.Context, userUUID uuid.UUID, preferences UserTimePreferences) error {
	ctx, span := tracer.Start(ctx, "subscription_service.set_time_preferences")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", userUUID.String()),
		attribute.Int("preferences.weeks_count", len(preferences)),
	)

	// Transform map[int]map[types.DayOfWeek][]types.Lesson to []DBTimePreferences
	var dbPrefs []DBTimePreferences
	for weekNumber, weekPrefs := range preferences {
		for dayOfWeek, lessons := range weekPrefs {
			dbPrefs = append(dbPrefs, DBTimePreferences{
				UserUUID:   userUUID,
				WeekNumber: weekNumber,
				DayOfWeek:  dayOfWeek,
				Lessons:    lessons,
			})
		}
	}

	span.SetAttributes(attribute.Int("preferences.records_count", len(dbPrefs)))

	err := s.repo.SetTimePreferences(ctx, userUUID, dbPrefs)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "SetTimePreferences",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store time preferences in repository")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) GetTeacherPreferences(ctx context.Context, userUUID uuid.UUID) (UserTeacherPreferences, error) {
	ctx, span := tracer.Start(ctx, "subscription_service.get_teacher_preferences")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	dbPrefs, err := s.repo.GetTeacherPreferences(ctx, userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetTeacherPreferences",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve teacher preferences from repository")
		return nil, err
	}

	span.SetAttributes(attribute.Int("preferences.blacklisted_count", len(dbPrefs.BlacklistedTeachers)))
	span.SetStatus(codes.Ok, "")

	return dbPrefs.BlacklistedTeachers, nil
}

func (s *Service) SetTeacherPreferences(ctx context.Context, userUUID uuid.UUID, preferences UserTeacherPreferences) error {
	ctx, span := tracer.Start(ctx, "subscription_service.set_teacher_preferences")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", userUUID.String()),
		attribute.Int("preferences.blacklisted_count", len(preferences)),
	)

	dbPrefs := &DBTeacherPreferences{
		UserUUID:            userUUID,
		BlacklistedTeachers: preferences,
	}

	err := s.repo.SetTeacherPreferences(ctx, userUUID, dbPrefs)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "SetTeacherPreferences",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store teacher preferences in repository")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) GetMatchingSubscriptions(ctx context.Context, req *GetMatchingSubscriptionsReq) ([]GetMatchingSubscriptionsRes, error) {
	ctx, span := tracer.Start(ctx, "subscription_service.get_matching_subscriptions")
	defer span.End()

	span.SetAttributes(
		attribute.String("lab.type", string(req.LabType)),
		attribute.String("lab.topic", string(req.LabTopic)),
		attribute.Int("lab.number", req.LabNumber),
		attribute.Int("lab.auditorium", req.LabAuditorium),
		attribute.Int("available_slots.count", len(req.AvailableSlots)),
	)

	search := &DBSubscriptionSearch{
		LabType:        req.LabType,
		LabTopic:       req.LabTopic,
		LabNumber:      req.LabNumber,
		LabAuditorium:  req.LabAuditorium,
		AvailableSlots: req.AvailableSlots,
	}

	// Get matches from repository
	matches, err := s.repo.GetMatchingSubscriptionsBySlot(ctx, search)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetMatchingSubscriptions",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve matching subscriptions from repository")
		return nil, err
	}

	span.SetAttributes(attribute.Int("matches.before_deduplication", len(matches)))

	relevantMatches, err := s.deduplicator.Deduplicate(ctx, req, matches)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetMatchingSubscriptions",
			Step:      "Deduplication",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to deduplicate matching subscriptions")
		return nil, err
	}

	result := make([]GetMatchingSubscriptionsRes, len(relevantMatches))
	for i, match := range relevantMatches {
		result[i] = GetMatchingSubscriptionsRes{
			UserUUID:                   match.UserUUID,
			SubscriptionUUID:           match.SubscriptionUUID,
			AutoEnroll:                 match.AutoEnroll,
			SuccessfulSubscriptions:    match.SuccessfulSubscriptions,
			LastSuccessfulSubscription: match.LastSuccessfulSubscription,
			MatchingTimeslots:          match.MatchingTimeslots,
		}
	}

	span.SetAttributes(
		attribute.Int("matches.after_deduplication", len(result)),
		attribute.Int("matches.filtered_out", len(matches)-len(result)),
	)
	span.SetStatus(codes.Ok, "")
	return result, nil
}
