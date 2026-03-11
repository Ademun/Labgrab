package subscription

import (
	"context"
	"fmt"
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/types"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo         *Repo
	deduplicator *Deduplicator
}

func NewService(repo *Repo, deduplicator *Deduplicator) *Service {
	return &Service{repo: repo, deduplicator: deduplicator}
}

func (s *Service) CreateSubscription(ctx context.Context, req *CreateSubscriptionReq) (uuid.UUID, error) {
	if err := req.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("subscription service: create subscription: validate request: %w", err)
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
		return uuid.Nil, fmt.Errorf("subscription service: create subscription: repository call: %w", err)
	}

	return subscriptionUUID, nil
}

func (s *Service) GetSubscription(ctx context.Context, subscriptionUUID uuid.UUID) (*GetSubscriptionRes, error) {
	sub, err := s.repo.GetSubscription(ctx, subscriptionUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription service: get subscription: repository call: %w", err)
	}

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
	subs, err := s.repo.GetSubscriptions(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription service: get subscriptions: repository call: %w", err)
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

	return result, nil
}

func (s *Service) UpdateSubscription(ctx context.Context, req *UpdateSubscriptionDataReq) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("subscription service: update subscription: validate request: %w", err)
	}

	err := s.repo.UpdateSubscription(ctx, &DBSubscription{
		SubscriptionUUID: req.SubscriptionUUID,
		LabType:          req.LabType,
		LabTopic:         req.LabTopic,
		LabNumber:        req.LabNumber,
		LabAuditorium:    req.LabAuditorium,
		Status:           req.Status,
		AutoEnroll:       req.AutoEnroll,
		AnyDate:          req.AnyDate,
		UserUUID:         req.UserUUID,
	})
	if err != nil {
		return fmt.Errorf("subscription service: update subscription: repository call: %w", err)
	}

	return nil
}

func (s *Service) GetTimeRestrictions(ctx context.Context, userUUID uuid.UUID) (UserTimeRestrictions, error) {
	dbRest, err := s.repo.GetTimeRestrictions(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription service: get time restrictions: repository call: %w", err)
	}

	userRest := make(UserTimeRestrictions)
	for _, pref := range dbRest {
		if _, exists := userRest[pref.WeekNumber]; !exists {
			userRest[pref.WeekNumber] = make(map[types.DayOfWeek][]domain.Lesson)
		}
		userRest[pref.WeekNumber][pref.DayOfWeek] = pref.Lessons
	}

	return userRest, nil
}

func (s *Service) SetTimeRestrictions(ctx context.Context, userUUID uuid.UUID, restrictions UserTimeRestrictions) error {
	var dbRest []DBTimeRestrictions
	for weekNumber, weekRest := range restrictions {
		for dayOfWeek, lessons := range weekRest {
			dbRest = append(dbRest, DBTimeRestrictions{
				UserUUID:   userUUID,
				WeekNumber: weekNumber,
				DayOfWeek:  dayOfWeek,
				Lessons:    lessons,
			})
		}
	}

	if err := s.repo.SetTimeRestrictions(ctx, userUUID, dbRest); err != nil {
		return fmt.Errorf("subscription service: set time restrictions: repository call: %w", err)
	}

	return nil
}

func (s *Service) GetTeacherPreferences(ctx context.Context, userUUID uuid.UUID) (UserTeacherPreferences, error) {
	dbPrefs, err := s.repo.GetTeacherPreferences(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription service: get teacher preferences: repository call: %w", err)
	}

	return dbPrefs.BlacklistedTeachers, nil
}

func (s *Service) SetTeacherPreferences(ctx context.Context, userUUID uuid.UUID, preferences UserTeacherPreferences) error {
	if err := s.repo.SetTeacherPreferences(ctx, userUUID, &DBTeacherPreferences{
		UserUUID:            userUUID,
		BlacklistedTeachers: preferences,
	}); err != nil {
		return fmt.Errorf("subscription service: set teacher preferences: repository call: %w", err)
	}

	return nil
}

func (s *Service) GetMatchingSubscriptions(ctx context.Context, req *GetMatchingSubscriptionsReq) ([]GetMatchingSubscriptionsRes, error) {
	relevantSlots := make(domain.Schedule)
	for date, data := range req.Schedule {
		if date.Sub(time.Now()).Hours() >= 48 {
			relevantSlots[date] = data
		}
	}

	matches, err := s.repo.GetMatchingSubscriptionsBySlot(ctx, &DBSubscriptionSearch{
		LabType:        req.Type,
		LabTopic:       req.Topic,
		LabNumber:      req.Number,
		LabAuditorium:  req.Auditorium,
		AvailableSlots: relevantSlots,
	})
	if err != nil {
		return nil, fmt.Errorf("subscription service: get matching subscriptions: repository call: %w", err)
	}

	relevantMatches, err := s.deduplicator.Deduplicate(ctx, req, matches)
	if err != nil {
		return nil, fmt.Errorf("subscription service: get matching subscriptions: deduplicate: %w", err)
	}

	result := make([]GetMatchingSubscriptionsRes, len(relevantMatches))
	for i, match := range relevantMatches {
		result[i] = GetMatchingSubscriptionsRes{
			UserUUID:                   match.UserUUID,
			SubscriptionUUID:           match.SubscriptionUUID,
			AutoEnroll:                 match.AutoEnroll,
			AnyDate:                    match.AnyDate,
			SuccessfulSubscriptions:    match.SuccessfulSubscriptions,
			LastSuccessfulSubscription: match.LastSuccessfulSubscription,
			Schedule:                   match.MatchingTimeslots,
		}
	}

	return result, nil
}

func (s *Service) CloseSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	if err := s.repo.CloseSubscription(ctx, subscriptionUUID); err != nil {
		return fmt.Errorf("subscription service: close subscription: repository call: %w", err)
	}
	return nil
}

func (s *Service) DeleteSubscriptions(ctx context.Context, req *DeleteSubscriptionsReq) error {
	if err := s.repo.DeleteSubscriptions(ctx, req.UserUUID, req.Tx); err != nil {
		return fmt.Errorf("subscription service: delete subscriptions: repository call: %w", err)
	}
	return nil
}
