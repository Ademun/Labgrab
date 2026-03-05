package subscription

import (
	"context"
	"encoding/json"
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/errors"
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

func (r *Repo) GetTimePreferences(ctx context.Context, userUUID uuid.UUID) ([]DBTimePreferences, error) {
	query, args, err := r.sq.Select(
		"user_uuid",
		"week_number",
		"day_of_week",
		"lessons",
	).
		From("subscription_service.time_preferences").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetTimePreferences",
			Step:      "Query setup",
			Err:       err,
		}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetTimePreferences",
			Step:      "Query execution",
			Err:       err,
		}
	}
	defer rows.Close()

	var preferences []DBTimePreferences
	for rows.Next() {
		var pref DBTimePreferences
		err = rows.Scan(
			&pref.UserUUID,
			&pref.WeekNumber,
			&pref.DayOfWeek,
			&pref.Lessons,
		)
		if err != nil {
			return nil, &errors.ErrDBProcedure{
				Procedure: "GetTimePreferences",
				Step:      "Row scanning",
				Err:       err,
			}
		}
		preferences = append(preferences, pref)
	}

	if err = rows.Err(); err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetTimePreferences",
			Step:      "Row error check",
			Err:       err,
		}
	}

	return preferences, nil
}

func (r *Repo) SetTimePreferences(ctx context.Context, userUUID uuid.UUID, preferences []DBTimePreferences) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "SetTimePreferences",
			Step:      "Begin transaction",
			Err:       err,
		}
	}
	defer tx.Rollback(ctx)

	deleteQuery, deleteArgs, err := r.sq.Delete("subscription_service.time_preferences").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "SetTimePreferences",
			Step:      "Delete query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, deleteQuery, deleteArgs...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "SetTimePreferences",
			Step:      "Delete execution",
			Err:       err,
		}
	}

	if len(preferences) > 0 {
		insertBuilder := r.sq.Insert("subscription_service.time_preferences").
			Columns("user_uuid", "week_number", "day_of_week", "lessons")

		for _, pref := range preferences {
			insertBuilder = insertBuilder.Values(pref.UserUUID, pref.WeekNumber, pref.DayOfWeek, pref.Lessons)
		}

		insertQuery, insertArgs, err := insertBuilder.ToSql()
		if err != nil {
			return &errors.ErrDBProcedure{
				Procedure: "SetTimePreferences",
				Step:      "Insert query setup",
				Err:       err,
			}
		}

		_, err = tx.Exec(ctx, insertQuery, insertArgs...)
		if err != nil {
			return &errors.ErrDBProcedure{
				Procedure: "SetTimePreferences",
				Step:      "Insert execution",
				Err:       err,
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "SetTimePreferences",
			Step:      "Commit transaction",
			Err:       err,
		}
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
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetTeacherPreferences",
			Step:      "Query setup",
			Err:       err,
		}
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
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetTeacherPreferences",
			Step:      "Query execution",
			Err:       err,
		}
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
		return &errors.ErrDBProcedure{
			Procedure: "SetTeacherPreferences",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "SetTeacherPreferences",
			Step:      "Query execution",
			Err:       err,
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
        auto_enroll,
        any_date,
        successful_subscriptions,
        last_successful_subscription,
        time,
        jsonb_object_agg(lesson::text, teachers ORDER BY lesson) as lessons_map
    FROM matching_subscriptions
    GROUP BY user_uuid, subscription_uuid, auto_enroll, any_date, successful_subscriptions, last_successful_subscription, time
),

candidates AS (
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
),

locked AS (
    UPDATE subscription_service.subscriptions s
    SET 
        locked_until = NOW() + INTERVAL '10 minutes',
        locked_by = $6
    FROM candidates c
    WHERE s.subscription_uuid = c.subscription_uuid
      AND (s.locked_until IS NULL OR s.locked_until < NOW())
    RETURNING s.subscription_uuid
)

SELECT c.*
FROM candidates c
INNER JOIN locked l ON c.subscription_uuid = l.subscription_uuid
`

	rows, err := r.pool.Query(ctx, query,
		search.LabType,
		search.LabTopic,
		search.LabNumber,
		search.LabAuditorium,
		availableSlotsJSON,
		search.WorkerUUID,
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
			anyDate                    bool
			successfulSubscriptions    int
			lastSuccessfulSubscription *time.Time
			matchingTimeslotsJSON      []byte
		)

		err = rows.Scan(
			&userUUID,
			&subscriptionUUID,
			&autoEnroll,
			&anyDate,
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
			AnyDate:                    anyDate,
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
