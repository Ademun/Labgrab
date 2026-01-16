package subscription

import (
	"time"

	"github.com/google/uuid"
)

type LabType string

const (
	LabTypeDefence     LabType = "Defence"
	LabTypePerformance LabType = "Performance"
)

type LabTopic string

const (
	LabTopicVirtual     LabTopic = "Virtual"
	LabTopicElectricity LabTopic = "Electricity"
	LabTopicMechanics   LabTopic = "Mechanics"
)

type DayOfWeek string

const (
	DayMon DayOfWeek = "MON"
	DayTue DayOfWeek = "TUE"
	DayWed DayOfWeek = "WED"
	DayThu DayOfWeek = "THU"
	DayFri DayOfWeek = "FRI"
	DaySat DayOfWeek = "SAT"
	DaySun DayOfWeek = "SUN"
)

// DBSubscription subscription_service.subscriptions
type DBSubscription struct {
	SubscriptionUUID uuid.UUID  `db:"subscription_uuid"`
	LabType          LabType    `db:"lab_type"`
	LabTopic         LabTopic   `db:"lab_topic"`
	LabNumber        int        `db:"lab_number"`
	LabAuditorium    *int       `db:"lab_auditorium"` // Defence can happen in any auditorium
	CreatedAt        time.Time  `db:"created_at"`
	ClosedAt         *time.Time `db:"closed_at"`
	UserUUID         uuid.UUID  `db:"user_uuid"`
}

// DBTimePreferences subscription_service.time_preferences
type DBTimePreferences struct {
	DayOfWeek DayOfWeek `db:"day_of_week"`
	Lessons   []int     `db:"lessons"`
	UserUUID  uuid.UUID `db:"user_uuid"`
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
	TimePreferences            map[DayOfWeek][]int
	BlacklistedTeachers        []string
	SuccessfulSubscriptions    int
	LastSuccessfulSubscription *time.Time
	UserUUID                   uuid.UUID
}

type DBSubscriptionSearch struct {
	LabType        LabType
	LabTopic       LabTopic
	LabNumber      int
	LabAuditorium  int
	AvailableSlots map[DayOfWeek]map[int][]string
}

type DBSubscriptionMatchResult struct {
	UserUUID                   uuid.UUID
	SubscriptionUUID           uuid.UUID
	SuccessfulSubscriptions    int
	LastSuccessfulSubscription *time.Time
	MatchingTimeslots          map[DayOfWeek][]int
}

type CreateSubscriptionReq struct {
	UserUUID      uuid.UUID
	LabType       LabType
	LabTopic      LabTopic
	LabNumber     int
	LabAuditorium *int
	CreatedAt     time.Time
}

type CreateSubscriptionDataReq struct {
	UserUUID            uuid.UUID
	TimePreferences     map[DayOfWeek][]int
	BlacklistedTeachers []string
}

type UpdateSubscriptionDataReq struct {
	UserUUID         uuid.UUID
	SubscriptionUUID uuid.UUID
	LabType          LabType
	LabTopic         LabTopic
	LabNumber        int
	LabAuditorium    *int
}

type GetMatchingSubscriptionsReq struct {
	LabType        LabType
	LabTopic       LabTopic
	LabNumber      int
	LabAuditorium  int
	AvailableSlots map[DayOfWeek]map[int][]string
}

type GetSubscriptionRes struct {
	SubscriptionUUID uuid.UUID
	LabType          LabType
	LabTopic         LabTopic
	LabNumber        int
	LabAuditorium    *int
	CreatedAt        time.Time
	ClosedAt         *time.Time
}

type GetMatchingSubscriptionsRes struct {
	UserUUID                   uuid.UUID
	SubscriptionUUID           uuid.UUID
	SuccessfulSubscriptions    int
	LastSuccessfulSubscription *time.Time
	MatchingTimeslots          map[DayOfWeek][]int
}

type keyGenerationParams struct {
	subscriptionUUID uuid.UUID
	labType          LabType
	labTopic         LabTopic
	labNumber        int
	labAuditorium    int
	day              DayOfWeek
	lesson           int
}
