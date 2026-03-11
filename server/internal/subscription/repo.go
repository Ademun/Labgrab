package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"labgrab/internal/shared/domain"
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
		return uuid.Nil, fmt.Errorf("subscription repo: create subscription: build query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return uuid.Nil, fmt.Errorf("subscription repo: create subscription: exec query: %w", err)
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
		return nil, fmt.Errorf("subscription repo: get subscription: build query: %w", err)
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
		return nil, fmt.Errorf("subscription repo: get subscription: scan row: %w", err)
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
		return nil, fmt.Errorf("subscription repo: get subscriptions: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("subscription repo: get subscriptions: exec query: %w", err)
	}
	defer rows.Close()

	var subscriptions []DBSubscription
	for rows.Next() {
		var sub DBSubscription
		if err = rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("subscription repo: get subscriptions: scan rows: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("subscription repo: get subscriptions: rows error: %w", err)
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
		return fmt.Errorf("subscription repo: update subscription: build query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("subscription repo: update subscription: exec query: %w", err)
	}

	return nil
}

func (r *Repo) GetTimeRestrictions(ctx context.Context, userUUID uuid.UUID) ([]DBTimeRestrictions, error) {
	query, args, err := r.sq.Select(
		"user_uuid",
		"week_number",
		"day_of_week",
		"lessons",
	).
		From("subscription_service.time_restrictions").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("subscription repo: get time restrictions: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("subscription repo: get time preferences: exec query: %w", err)
	}
	defer rows.Close()

	var restrictions []DBTimeRestrictions
	for rows.Next() {
		var rest DBTimeRestrictions
		if err = rows.Scan(
			&rest.UserUUID,
			&rest.WeekNumber,
			&rest.DayOfWeek,
			&rest.Lessons,
		); err != nil {
			return nil, fmt.Errorf("subscription repo: get time restrictions: scan rows: %w", err)
		}
		restrictions = append(restrictions, rest)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("subscription repo: get time restrictions: rows error: %w", err)
	}

	return restrictions, nil
}

func (r *Repo) SetTimeRestrictions(ctx context.Context, userUUID uuid.UUID, preferences []DBTimeRestrictions) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("subscription repo: set time restrictions: begin tx: %w", err)
	}

	deleteQuery, deleteArgs, err := r.sq.Delete("subscription_service.time_restrictions").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("subscription repo: set time restrictions: build delete query: %w", err)
	}

	_, err = tx.Exec(ctx, deleteQuery, deleteArgs...)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("subscription repo: set time restrictions: exec delete query: %w", err)
	}

	if len(preferences) > 0 {
		insertBuilder := r.sq.Insert("subscription_service.time_restrictions").
			Columns("user_uuid", "week_number", "day_of_week", "lessons")

		for _, pref := range preferences {
			insertBuilder = insertBuilder.Values(pref.UserUUID, pref.WeekNumber, pref.DayOfWeek, pref.Lessons)
		}

		insertQuery, insertArgs, err := insertBuilder.ToSql()
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("subscription repo: set time restrictions: build insert query: %w", err)
		}

		_, err = tx.Exec(ctx, insertQuery, insertArgs...)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("subscription repo: set time restrictions: exec insert query: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("subscription repo: set time restrictions: commit tx: %w", err)
	}

	return nil
}

func (r *Repo) GetTeacherPreferences(ctx context.Context, userUUID uuid.UUID) (*DBTeacherPreferences, error) {
	query, args, err := r.sq.Select(
		"user_uuid",
		"blacklisted_teachers",
	).
		From("subscription_service.teacher_preferences").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("subscription repo: get teacher preferences: build query: %w", err)
	}

	var pref DBTeacherPreferences
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&pref.UserUUID,
		&pref.BlacklistedTeachers,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &DBTeacherPreferences{
				UserUUID:            userUUID,
				BlacklistedTeachers: []string{},
			}, nil
		}
		return nil, fmt.Errorf("subscription repo: get teacher preferences: scan row: %w", err)
	}

	return &pref, nil
}

func (r *Repo) SetTeacherPreferences(ctx context.Context, userUUID uuid.UUID, preferences *DBTeacherPreferences) error {
	query, args, err := r.sq.Insert("subscription_service.teacher_preferences").
		Columns("user_uuid", "blacklisted_teachers").
		Values(userUUID, preferences.BlacklistedTeachers).
		Suffix("ON CONFLICT (user_uuid) DO UPDATE SET blacklisted_teachers = EXCLUDED.blacklisted_teachers").
		ToSql()
	if err != nil {
		return fmt.Errorf("subscription repo: set teacher preferences: build query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("subscription repo: set teacher preferences: exec query: %w", err)
	}

	return nil
}

func (r *Repo) GetMatchingSubscriptionsBySlot(ctx context.Context, search *DBSubscriptionSearch) ([]DBSubscriptionMatchResult, error) {
	availableSlotsJSON, err := json.Marshal(search.AvailableSlots)
	if err != nil {
		return nil, fmt.Errorf("subscription repo: get matching subscriptions by slot: marshal available slots: %w", err)
	}

	query := `WITH available_slots_expanded AS (
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
        s.any_date,
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
                tr.day_of_week = ase.weekday::day_of_week
                AND ase.lesson = ANY(tr.lessons)
				AND (2 - ABS(EXTRACT(WEEK from ase.time::timestamp) % 2 - 1)) = tr.week_number
            ) as is_overlap
        FROM subscription_service.time_restrictions tr
        WHERE tr.user_uuid = s.user_uuid
        HAVING count(*) > 0 
    ) pref ON TRUE
    LEFT JOIN subscription_service.teacher_preferences teachp 
        ON s.user_uuid = teachp.user_uuid
    WHERE s.lab_type = $1
      AND s.lab_topic = $2
      AND s.lab_number = $3
      AND (s.lab_auditorium = $4 OR (s.lab_auditorium IS NULL AND s.lab_type = 'Defence' AND $1 = 'Defence'))
      AND s.status = 'Active'
      AND (pref.has_any IS NULL OR pref.is_overlap IS FALSE OR s.any_date IS TRUE)
      AND (teachp.user_uuid IS NULL OR NOT (ase.teachers ?| teachp.blacklisted_teachers))
),
grouped_by_time AS (
    SELECT 
        user_uuid,
        subscription_uuid,
        auto_enroll,
        any_date,
        successful_subscriptions,
        last_successful_subscription,
        time,
        jsonb_object_agg(lesson::text, teachers ORDER BY lesson) as lessons_map
    FROM matching_subscriptions
    GROUP BY user_uuid, subscription_uuid, auto_enroll, any_date, successful_subscriptions, last_successful_subscription, time
)
SELECT 
    user_uuid,
    subscription_uuid,
    auto_enroll,
    any_date,
    successful_subscriptions,
    last_successful_subscription,
    jsonb_object_agg(time, lessons_map) as matching_timeslots
FROM grouped_by_time
GROUP BY user_uuid, subscription_uuid, auto_enroll, any_date, successful_subscriptions, last_successful_subscription
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
		return nil, fmt.Errorf("subscription repo: get matching subscriptions by slot: exec query: %w", err)
	}
	defer rows.Close()

	var results []DBSubscriptionMatchResult
	for rows.Next() {
		var (
			userUUID                   uuid.UUID
			subscriptionUUID           uuid.UUID
			autoEnroll                 bool
			anyDate                    bool
			successfulSubscriptions    int
			lastSuccessfulSubscription *time.Time
			matchingTimeslotsJSON      []byte
		)

		if err = rows.Scan(
			&userUUID,
			&subscriptionUUID,
			&autoEnroll,
			&anyDate,
			&successfulSubscriptions,
			&lastSuccessfulSubscription,
			&matchingTimeslotsJSON,
		); err != nil {
			return nil, fmt.Errorf("subscription repo: get matching subscriptions by slot: scan rows: %w", err)
		}

		var matchingTimeslots domain.Schedule
		if err = json.Unmarshal(matchingTimeslotsJSON, &matchingTimeslots); err != nil {
			return nil, fmt.Errorf("subscription repo: get matching subscriptions by slot: unmarshal timeslots: %w", err)
		}

		results = append(results, DBSubscriptionMatchResult{
			UserUUID:                   userUUID,
			SubscriptionUUID:           subscriptionUUID,
			AutoEnroll:                 autoEnroll,
			AnyDate:                    anyDate,
			SuccessfulSubscriptions:    successfulSubscriptions,
			LastSuccessfulSubscription: lastSuccessfulSubscription,
			MatchingTimeslots:          matchingTimeslots,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("subscription repo: get matching subscriptions by slot: rows error: %w", err)
	}

	return results, nil
}

func (r *Repo) CloseSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	query, args, err := r.sq.Update("subscription_service.subscriptions").
		Set("status", StatusClosed).
		Where(squirrel.Eq{"subscription_uuid": subscriptionUUID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("subscription repo: close subscription: build query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("subscription repo: close subscription: exec query: %w", err)
	}

	return nil
}

func (r *Repo) DeleteSubscriptions(ctx context.Context, userUUID uuid.UUID, tx pgx.Tx) error {
	query, args, err := r.sq.Delete("subscription_service.subscriptions").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("subscription repo: delete subscriptions: build query: %w", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("subscription repo: delete subscriptions: exec query: %w", err)
	}
	return nil
}
