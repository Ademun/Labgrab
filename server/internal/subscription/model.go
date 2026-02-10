package subscription

import (
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/errors"
	"labgrab/internal/shared/types"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Status string

const (
	StatusActive Status = "Active"
	StatusPaused Status = "Paused"
	StatusClosed Status = "Closed"
)

// DBSubscription subscription_service.subscriptions
type DBSubscription struct {
	SubscriptionUUID uuid.UUID       `db:"subscription_uuid"`
	LabType          domain.LabType  `db:"lab_type"`
	LabTopic         domain.LabTopic `db:"lab_topic"`
	LabNumber        int             `db:"lab_number"`
	LabAuditorium    *int            `db:"lab_auditorium"` // Defence can happen in any auditorium
	Status           Status          `db:"status"`
	AutoEnroll       bool            `db:"auto_enroll"`
	AnyDate          bool            `db:"any_date"`
	CreatedAt        time.Time       `db:"created_at"`
	ClosedAt         *time.Time      `db:"closed_at"`
	UserUUID         uuid.UUID       `db:"user_uuid"`
}

// DBTimePreferences subscription_service.time_preferences
type DBTimePreferences struct {
	DayOfWeek types.DayOfWeek `db:"day_of_week"`
	Lessons   []domain.Lesson `db:"lessons"`
	UserUUID  uuid.UUID       `db:"user_uuid"`
}

// DBTeacherPreferences subscription_service.teacher_preferences
type DBTeacherPreferences struct {
	BlacklistedTeachers []string  `db:"blacklisted_teachers"`
	UserUUID            uuid.UUID `db:"user_uuid"`
}

// DBDetails subscription_service.details
type DBDetails struct {
	SuccessfulSubscriptions    int        `db:"successful_subscriptions"`
	LastSuccessfulSubscription *time.Time `db:"last_successful_subscription"`
	UserUUID                   uuid.UUID  `db:"user_uuid"`
}

type DBUserSubscriptionData struct {
	TimePreferences            map[types.DayOfWeek][]domain.Lesson
	BlacklistedTeachers        []string
	SuccessfulSubscriptions    int
	LastSuccessfulSubscription *time.Time
	UserUUID                   uuid.UUID
}

type DBSubscriptionSearch struct {
	LabType        domain.LabType
	LabTopic       domain.LabTopic
	LabNumber      int
	LabAuditorium  int
	AvailableSlots domain.Schedule
}

type DBSubscriptionMatchResult struct {
	UserUUID                   uuid.UUID
	SubscriptionUUID           uuid.UUID
	SuccessfulSubscriptions    int
	LastSuccessfulSubscription *time.Time
	MatchingTimeslots          domain.Schedule
}

type CreateSubscriptionReq struct {
	UserUUID      uuid.UUID
	LabType       domain.LabType
	LabTopic      domain.LabTopic
	LabNumber     int
	LabAuditorium *int
	AutoEnroll    bool `db:"auto_enroll"`
	AnyDate       bool `db:"any_date"`
	CreatedAt     time.Time
}

func (r CreateSubscriptionReq) Validate() error {
	err := errors.NewValidationError()
	if r.LabType == domain.LabTypePerformance && r.LabAuditorium == nil {
		err.Add("lab_type & lab_auditorium", "If lab type is equal to 'Performance' lab auditorium should be provided")
	}
	if r.LabType == domain.LabTypeDefence && r.LabAuditorium != nil {
		err.Add("lab_type & lab_auditorium", "If lab type is equal to 'Defence' lab auditorium should not be provided")
	}
	if err.HasErrors() {
		return err
	}
	return nil
}

type CreateSubscriptionDataReq struct {
	UserUUID            uuid.UUID
	TimePreferences     map[types.DayOfWeek][]domain.Lesson
	BlacklistedTeachers []string
	Tx                  pgx.Tx
}

type UpdateSubscriptionDataReq struct {
	UserUUID         uuid.UUID
	SubscriptionUUID uuid.UUID
	LabType          domain.LabType
	LabTopic         domain.LabTopic
	LabNumber        int
	LabAuditorium    *int
	Status           Status `db:"status"`
	AutoEnroll       bool   `db:"auto_enroll"`
	AnyDate          bool   `db:"any_date"`
}

func (r UpdateSubscriptionDataReq) Validate() error {
	err := errors.NewValidationError()
	if r.LabType == domain.LabTypePerformance && r.LabAuditorium == nil {
		err.Add("lab_type & lab_auditorium", "If lab type is equal to 'Performance' lab auditorium should be provided")
	}
	if r.LabType == domain.LabTypeDefence && r.LabAuditorium != nil {
		err.Add("lab_type & lab_auditorium", "If lab type is equal to 'Defence' lab auditorium should not be provided")
	}
	if err.HasErrors() {
		return err
	}
	return nil
}

type GetMatchingSubscriptionsReq struct {
	LabType        domain.LabType
	LabTopic       domain.LabTopic
	LabNumber      int
	LabAuditorium  int
	AvailableSlots domain.Schedule
}

type GetSubscriptionRes struct {
	SubscriptionUUID uuid.UUID
	LabType          domain.LabType
	LabTopic         domain.LabTopic
	LabNumber        int
	LabAuditorium    *int
	Status           Status `db:"status"`
	AutoEnroll       bool   `db:"auto_enroll"`
	AnyDate          bool   `db:"any_date"`
	CreatedAt        time.Time
	ClosedAt         *time.Time
}

type GetMatchingSubscriptionsRes struct {
	UserUUID                   uuid.UUID
	SubscriptionUUID           uuid.UUID
	SuccessfulSubscriptions    int
	LastSuccessfulSubscription *time.Time
	MatchingTimeslots          domain.Schedule
}

type keyGenerationParams struct {
	subscriptionUUID uuid.UUID
	labType          domain.LabType
	labTopic         domain.LabTopic
	labNumber        int
	labAuditorium    int
	time             time.Time
	lesson           domain.Lesson
}
