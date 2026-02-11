package subscription

import (
	"context"
	"encoding/json"
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/errors"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
	sq   squirrel.StatementBuilderType
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool, sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *Repo) CreateSubscription(ctx context.Context, sub *DBSubscription) (uuid.UUID, error) {
	subscriptionUUID := uuid.New()

	query, args, err := r.sq.Insert("subscription_service.subscriptions").
		Columns("subscription_uuid", "lab_type", "lab_topic", "lab_number", "lab_auditorium", "status", "auto_enroll", "any_date", "created_at", "user_uuid").
		Values(subscriptionUUID, sub.LabType, sub.LabTopic, sub.LabNumber, sub.LabAuditorium, StatusActive, sub.AutoEnroll, sub.AnyDate, sub.CreatedAt, sub.UserUUID).
		ToSql()
	if err != nil {
		return uuid.Nil, &errors.ErrDBProcedure{
			Procedure: "CreateSubscription",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return uuid.Nil, &errors.ErrDBProcedure{
			Procedure: "CreateSubscription",
			Step:      "Query execution",
			Err:       err,
		}
	}

	return subscriptionUUID, nil
}

func (r *Repo) GetSubscription(ctx context.Context, subscriptionUUID uuid.UUID) (*DBSubscription, error) {
	query, args, err := r.sq.Select(
		"subscription_uuid",
		"lab_type",
		"lab_topic",
		"lab_number",
		"lab_auditorium",
		"status",
		"auto_enroll",
		"any_date",
		"created_at",
		"closed_at",
		"user_uuid",
	).
		From("subscription_service.subscriptions").
		Where(squirrel.Eq{"subscription_uuid": subscriptionUUID}).
		ToSql()
	if err != nil {
		return nil,
			&errors.ErrDBProcedure{
				Procedure: "GetSubscription",
				Step:      "Query setup",
				Err:       err,
			}
	}

	var sub DBSubscription
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&sub.SubscriptionUUID,
		&sub.LabType,
		&sub.LabTopic,
		&sub.LabNumber,
		&sub.LabAuditorium,
		&sub.Status,
		&sub.AutoEnroll,
		&sub.AnyDate,
		&sub.CreatedAt,
		&sub.ClosedAt,
		&sub.UserUUID,
	)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetSubscription",
			Step:      "Query execution",
			Err:       err,
		}
	}

	return &sub, nil
}

func (r *Repo) GetSubscriptions(ctx context.Context, userUUID uuid.UUID) ([]DBSubscription, error) {
	query, args, err := r.sq.Select(
		"subscription_uuid",
		"lab_type",
		"lab_topic",
		"lab_number",
		"lab_auditorium",
		"status",
		"auto_enroll",
		"any_date",
		"created_at",
		"closed_at",
		"user_uuid",
	).
		From("subscription_service.subscriptions").
		Where(squirrel.And{squirrel.Eq{"user_uuid": userUUID}, squirrel.NotEq{"status": StatusClosed}}).
		ToSql()
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetSubscriptions",
			Step:      "Query setup",
			Err:       err,
		}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetSubscriptions",
			Step:      "Query execution",
			Err:       err,
		}
	}
	defer rows.Close()

	var subscriptions []DBSubscription
	for rows.Next() {
		var sub DBSubscription
		err = rows.Scan(
			&sub.SubscriptionUUID,
			&sub.LabType,
			&sub.LabTopic,
			&sub.LabNumber,
			&sub.LabAuditorium,
			&sub.Status,
			&sub.AutoEnroll,
			&sub.AnyDate,
			&sub.CreatedAt,
			&sub.ClosedAt,
			&sub.UserUUID,
		)
		if err != nil {
			return nil, &errors.ErrDBProcedure{
				Procedure: "GetSubscriptions",
				Step:      "Row scanning",
				Err:       err,
			}
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetSubscriptions",
			Step:      "Row error check",
			Err:       err,
		}
	}

	return subscriptions, nil
}

func (r *Repo) UpdateSubscription(ctx context.Context, sub *DBSubscription) error {
	query, args, err := r.sq.Update("subscription_service.subscriptions").
		Set("lab_type", sub.LabType).
		Set("lab_topic", sub.LabTopic).
		Set("lab_number", sub.LabNumber).
		Set("lab_auditorium", sub.LabAuditorium).
		Set("status", sub.Status).
		Set("auto_enroll", sub.AutoEnroll).
		Set("any_date", sub.AnyDate).
		Where(squirrel.Eq{"subscription_uuid": sub.SubscriptionUUID}).
		ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "UpdateSubscription",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "UpdateSubscription",
			Step:      "Query execution",
			Err:       err,
		}
	}
	return nil
}

func (r *Repo) CreateSubscriptionData(ctx context.Context, data *DBUserSubscriptionData, tx pgx.Tx) error {
	detailsQuery, detailsArgs, err := r.sq.Insert("subscription_service.details").
		Columns("successful_subscriptions", "last_successful_subscription", "user_uuid").
		Values(data.SuccessfulSubscriptions, data.LastSuccessfulSubscription, data.UserUUID).
		ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CreateSubscriptionData",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, detailsQuery, detailsArgs...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CreateSubscriptionData",
			Step:      "Query execution",
			Err:       err,
		}
	}

	teacherQuery, teacherArgs, err := r.sq.Insert("subscription_service.teacher_preferences").
		Columns("blacklisted_teachers", "user_uuid").
		Values(data.BlacklistedTeachers, data.UserUUID).
		ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CreateSubscriptionData",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, teacherQuery, teacherArgs...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CreateSubscriptionData",
			Step:      "Query execution",
			Err:       err,
		}
	}

	for day, lessons := range data.TimePreferences {
		timeQuery, timeArgs, err := r.sq.Insert("subscription_service.time_preferences").
			Columns("day_of_week", "lessons", "user_uuid").
			Values(day, lessons, data.UserUUID).
			ToSql()
		if err != nil {
			return &errors.ErrDBProcedure{
				Procedure: "CreateSubscriptionData",
				Step:      "Query setup",
				Err:       err,
			}
		}

		_, err = tx.Exec(ctx, timeQuery, timeArgs...)
		if err != nil {
			return &errors.ErrDBProcedure{
				Procedure: "CreateSubscriptionData",
				Step:      "Query execution",
				Err:       err,
			}
		}
	}

	return nil
}

func (r *Repo) GetMatchingSubscriptionsBySlot(ctx context.Context, search *DBSubscriptionSearch) ([]DBSubscriptionMatchResult, error) {
	availableSlotsJSON, err := convertAvailableSlotsToJSON(search.AvailableSlots)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetMatchingSubscriptionsBySlot",
			Step:      "JSON conversion",
			Err:       err,
		}
	}

	query := `
WITH available_slots_expanded AS (
    SELECT 
        times.key AS time,
        TO_CHAR(times.key::timestamptz, 'DY') AS weekday,
        lessons.key::int AS lesson,
        lessons.value AS teachers
    FROM jsonb_each($5::jsonb) AS times,
         LATERAL jsonb_each(times.value) AS lessons
),

matching_subscriptions AS (
    SELECT 
        s.subscription_uuid,
		s.auto_enroll,
        s.user_uuid,
        d.successful_subscriptions,
        d.last_successful_subscription,
        ase.time,
        ase.weekday::day_of_week,
        ase.lesson,
        ase.teachers
    FROM subscription_service.subscriptions s
    INNER JOIN subscription_service.details d ON s.user_uuid = d.user_uuid
    CROSS JOIN available_slots_expanded ase
    LEFT JOIN LATERAL (
        SELECT 
            true as has_any,
            bool_or(
                tp.day_of_week = ase.weekday::day_of_week 
                AND ase.lesson = ANY(tp.lessons)
            ) as is_match
        FROM subscription_service.time_preferences tp
        WHERE tp.user_uuid = s.user_uuid
        HAVING count(*) > 0 
    ) pref ON TRUE
    LEFT JOIN subscription_service.teacher_preferences teachp 
        ON s.user_uuid = teachp.user_uuid
    WHERE s.lab_type = $1
      AND s.lab_topic = $2
      AND s.lab_number = $3
      AND (s.lab_auditorium = $4 OR (s.lab_auditorium IS NULL AND s.lab_type = 'Defence' AND $1 = 'Defence'))
      AND s.status = 'Active'
      AND (pref.has_any IS NULL OR pref.is_match IS TRUE OR s.any_date IS TRUE)
      AND (teachp.user_uuid IS NULL OR NOT (ase.teachers ?| teachp.blacklisted_teachers))
),
grouped_by_time AS (
    SELECT 
        user_uuid,
        subscription_uuid,
        successful_subscriptions,
        last_successful_subscription,
        time,
        jsonb_object_agg(lesson::text, teachers ORDER BY lesson) as lessons_map
    FROM matching_subscriptions
    GROUP BY user_uuid, subscription_uuid, successful_subscriptions, last_successful_subscription, time
)
SELECT 
    user_uuid,
    subscription_uuid,
	auto_enroll
    successful_subscriptions,
    last_successful_subscription,
    jsonb_object_agg(time, lessons_map) as matching_timeslots
FROM grouped_by_time
GROUP BY user_uuid, subscription_uuid, successful_subscriptions, last_successful_subscription
ORDER BY auto_enroll DESC,
    successful_subscriptions ASC,
    last_successful_subscription ASC NULLS FIRST
`

	rows, err := r.pool.Query(ctx, query,
		search.LabType,
		search.LabTopic,
		search.LabNumber,
		search.LabAuditorium,
		availableSlotsJSON,
	)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetMatchingSubscriptionsBySlot",
			Step:      "Query execution",
			Err:       err,
		}
	}
	defer rows.Close()

	var results []DBSubscriptionMatchResult

	for rows.Next() {
		var (
			userUUID                   uuid.UUID
			subscriptionUUID           uuid.UUID
			autoEnroll                 bool
			successfulSubscriptions    int
			lastSuccessfulSubscription *time.Time
			matchingTimeslotsJSON      []byte
		)

		err = rows.Scan(
			&userUUID,
			&subscriptionUUID,
			&autoEnroll,
			&successfulSubscriptions,
			&lastSuccessfulSubscription,
			&matchingTimeslotsJSON,
		)
		if err != nil {
			return nil, &errors.ErrDBProcedure{
				Procedure: "GetMatchingSubscriptionsBySlot",
				Step:      "Row scanning",
				Err:       err,
			}
		}

		matchingTimeslots, err := convertJSONToMatchingTimeslots(matchingTimeslotsJSON)
		if err != nil {
			return nil, &errors.ErrDBProcedure{
				Procedure: "GetMatchingSubscriptionsBySlot",
				Step:      "JSON conversion",
				Err:       err,
			}
		}

		results = append(results, DBSubscriptionMatchResult{
			UserUUID:                   userUUID,
			SubscriptionUUID:           subscriptionUUID,
			AutoEnroll:                 autoEnroll,
			SuccessfulSubscriptions:    successfulSubscriptions,
			LastSuccessfulSubscription: lastSuccessfulSubscription,
			MatchingTimeslots:          matchingTimeslots,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetMatchingSubscriptionsBySlot",
			Step:      "Row error check",
			Err:       err,
		}
	}

	for _, result := range results {
		slog.Info("result", result)
	}

	return results, nil
}

func convertAvailableSlotsToJSON(slots domain.Schedule) ([]byte, error) {
	return json.Marshal(slots)
}

func convertJSONToMatchingTimeslots(data []byte) (domain.Schedule, error) {
	var result domain.Schedule
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
