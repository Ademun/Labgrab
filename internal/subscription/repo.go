package subscription

import (
	"context"
	"encoding/json"
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
	subscriptionUUID, err := uuid.NewUUID()
	if err != nil {
		return uuid.Nil, &errors.ErrDBProcedure{
			Procedure: "CreateSubscription",
			Step:      "UUID generation",
			Err:       err,
		}
	}
	query, args, err := r.sq.Insert("subscription_service.subscriptions").
		Columns("subscription_uuid", "lab_type", "lab_topic", "lab_number", "lab_auditorium", "created_at", "user_uuid").
		Values(subscriptionUUID, sub.LabType, sub.LabTopic, sub.LabNumber, sub.LabAuditorium, sub.CreatedAt, sub.UserUUID).
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

	return subscriptionUUID, err
}

func (r *Repo) GetSubscription(ctx context.Context, subscriptionUUID uuid.UUID) (*DBSubscription, error) {
	query, args, err := r.sq.Select(
		"subscription_uuid",
		"lab_type",
		"lab_topic",
		"lab_number",
		"lab_auditorium",
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
		"created_at",
		"closed_at",
		"user_uuid",
	).
		From("subscription_service.subscriptions").
		Where(squirrel.Eq{"user_uuid": userUUID}).
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

func (r *Repo) CloseSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	query, args, err := r.sq.Update("subscription_service.subscriptions").
		Set("closed_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"subscription_uuid": subscriptionUUID}).
		ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CloseSubscription",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CloseSubscription",
			Step:      "Query execution",
			Err:       err,
		}
	}
	return nil
}

func (r *Repo) RestoreSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	query, args, err := r.sq.Update("subscription_service.subscriptions").
		Set("closed_at", nil).
		Where(squirrel.Eq{"subscription_uuid": subscriptionUUID}).
		ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "RestoreSubscription",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "RestoreSubscription",
			Step:      "Query execution",
			Err:       err,
		}
	}
	return nil
}

func (r *Repo) DeleteSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	query, args, err := r.sq.Delete("subscription_service.subscriptions").
		Where(squirrel.Eq{"subscription_uuid": subscriptionUUID}).
		ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "DeleteSubscription",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "DeleteSubscription",
			Step:      "Query execution",
			Err:       err,
		}
	}
	return nil
}

func (r *Repo) CreateSubscriptionData(ctx context.Context, tx pgx.Tx, data *DBUserSubscriptionData) error {
	detailsQuery, detailsArgs, err := r.sq.Insert("subscription_service.details").
		Columns("successful_subscriptions", "last_successful_subscription", "user_uuid").
		Values(data.SuccessfulSubscriptions, data.LastSuccessfulSubscription, data.UserUUID).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "CreateSubscriptionData",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, detailsQuery, detailsArgs...)
	if err != nil {
		tx.Rollback(ctx)
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
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "CreateSubscriptionData",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, teacherQuery, teacherArgs...)
	if err != nil {
		tx.Rollback(ctx)
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
			tx.Rollback(ctx)
			return &errors.ErrDBProcedure{
				Procedure: "CreateSubscriptionData",
				Step:      "Query setup",
				Err:       err,
			}
		}

		_, err = tx.Exec(ctx, timeQuery, timeArgs...)
		if err != nil {
			tx.Rollback(ctx)
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
        days.key::text AS day_of_week,
        lessons.key::int AS lesson,
        lessons.value AS teachers
    FROM jsonb_each($5::jsonb) AS days,
         LATERAL jsonb_each(days.value) AS lessons
),
matching_subscriptions AS (
    SELECT 
        s.subscription_uuid,
        s.user_uuid,
        d.successful_subscriptions,
        d.last_successful_subscription,
        ase.day_of_week::day_of_week,
        ase.lesson
    FROM subscription_service.subscriptions s
    INNER JOIN subscription_service.details d ON s.user_uuid = d.user_uuid
    CROSS JOIN available_slots_expanded ase
    INNER JOIN subscription_service.time_preferences tp 
        ON s.user_uuid = tp.user_uuid 
        AND tp.day_of_week = ase.day_of_week::day_of_week
        AND ase.lesson = ANY(tp.lessons)
    INNER JOIN subscription_service.teacher_preferences teachp 
        ON s.user_uuid = teachp.user_uuid
    WHERE s.lab_type = $1
      AND s.lab_topic = $2
      AND s.lab_number = $3
      AND (s.lab_auditorium IS NULL OR s.lab_auditorium = $4)
      AND s.closed_at IS NULL
      AND EXISTS (
          SELECT 1 
          FROM jsonb_array_elements_text(ase.teachers) teacher
          WHERE teacher != ALL(teachp.blacklisted_teachers)
      )
),
grouped_by_day AS (
    SELECT 
        user_uuid,
        subscription_uuid,
        successful_subscriptions,
        last_successful_subscription,
        day_of_week,
        jsonb_agg(DISTINCT lesson ORDER BY lesson) as lessons_array
    FROM matching_subscriptions
    GROUP BY user_uuid, subscription_uuid, successful_subscriptions, last_successful_subscription, day_of_week
)
SELECT 
    user_uuid,
    subscription_uuid,
    successful_subscriptions,
    last_successful_subscription,
    jsonb_object_agg(day_of_week, lessons_array) as matching_timeslots
FROM grouped_by_day
GROUP BY user_uuid, subscription_uuid, successful_subscriptions, last_successful_subscription
ORDER BY 
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
			successfulSubscriptions    int
			lastSuccessfulSubscription *time.Time
			matchingTimeslotsJSON      []byte
		)

		err = rows.Scan(
			&userUUID,
			&subscriptionUUID,
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

func convertAvailableSlotsToJSON(slots map[DayOfWeek]map[int][]string) ([]byte, error) {
	return json.Marshal(slots)
}

func convertJSONToMatchingTimeslots(data []byte) (map[DayOfWeek][]int, error) {
	var raw map[string][]int
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make(map[DayOfWeek][]int)
	for day, lessons := range raw {
		result[DayOfWeek(day)] = lessons
	}

	return result, nil
}
