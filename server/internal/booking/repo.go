package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"labgrab/internal/shared/domain"

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

func (r *Repo) LoadBookings(ctx context.Context, userUUID uuid.UUID, data []DBBooking) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("booking repo: load bookings: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	query, args, err := r.sq.Delete("booking_service.bookings").Where(squirrel.Eq{"user_uuid": userUUID}).ToSql()
	if err != nil {
		return fmt.Errorf("booking repo: load bookings: build delete query: %w", err)
	}

	if _, err = tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("booking repo: load bookings: exec delete query: %w", err)
	}

	builder := r.sq.Insert("booking_service.bookings").
		Columns("booking_id", "type", "topic", "number", "auditorium", "spot", "lesson", "start_time", "end_time", "status", "user_uuid")

	for _, b := range data {
		builder = builder.Values(b.BookingID, b.Type, b.Topic, b.Number, b.Auditorium, b.Spot, b.Lesson, b.Start, b.End, b.Status, b.UserUUID)
	}

	query, args, err = builder.ToSql()
	if err != nil {
		return fmt.Errorf("booking repo: load bookings: build insert query: %w", err)
	}

	if _, err = tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("booking repo: load bookings: exec insert query: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("booking repo: load bookings: commit tx: %w", err)
	}

	return nil
}

func (r *Repo) GetBookings(ctx context.Context, userUUID uuid.UUID) ([]DBBooking, error) {
	query, args, err := r.sq.Select(
		"booking_id",
		"type",
		"topic",
		"number",
		"auditorium",
		"spot",
		"lesson",
		"start_time",
		"end_time",
		"status",
		"user_uuid",
	).
		From("booking_service.bookings").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("booking repo: get bookings: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("booking repo: get bookings: exec query: %w", err)
	}
	defer rows.Close()

	var bookings []DBBooking
	for rows.Next() {
		var item DBBooking
		if err := rows.Scan(
			&item.BookingID,
			&item.Type,
			&item.Topic,
			&item.Number,
			&item.Auditorium,
			&item.Spot,
			&item.Lesson,
			&item.Start,
			&item.End,
			&item.Status,
			&item.UserUUID,
		); err != nil {
			return nil, fmt.Errorf("booking repo: get bookings: scan row: %w", err)
		}
		bookings = append(bookings, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("booking repo: get bookings: rows error: %w", err)
	}

	return bookings, nil
}

func (r *Repo) DeleteBookings(ctx context.Context, userUUID uuid.UUID, tx pgx.Tx) error {
	query, args, err := r.sq.Delete("booking_service.bookings").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("booking repo: delete bookings: exec delete query: %w", err)
	}

	if _, err = tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("booking repo: delete bookings: exec delete query: %w", err)
	}
	return nil
}

func (r *Repo) FilterSchedule(ctx context.Context, filter *DBSlotFilter) (domain.Schedule, error) {
	slotsJSON, err := convertScheduleToJSON(filter.Schedule)
	if err != nil {
		return nil, fmt.Errorf("booking repo: filter schedule: convert schedule to json: %w", err)
	}

	query := `
WITH already_booked AS (
    SELECT EXISTS(
        SELECT 1
        FROM booking_service.bookings
        WHERE user_uuid = $1
          AND status    = 'Open'
          AND type      = $3::lab_type
          AND topic     = $4::lab_topic
          AND number    = $5
    ) AS is_booked
),
schedule AS (
    SELECT
        times.key        AS time,
        lessons.key::int AS lesson,
        lessons.value    AS teachers
    FROM jsonb_each($2::jsonb) AS times,
         LATERAL jsonb_each(times.value) AS lessons
    WHERE NOT (SELECT is_booked FROM already_booked)
),
user_lessons AS (
    SELECT
        lesson,
        start_time::date AS lesson_date
    FROM booking_service.bookings
    WHERE user_uuid = $1
      AND status    = 'Open'
),
filtered_slots AS (
    SELECT
        sc.time,
        sc.lesson,
        sc.teachers
    FROM schedule sc
    WHERE NOT EXISTS (
        SELECT 1
        FROM user_lessons ul
        WHERE ABS(ul.lesson_date - sc.time::timestamp::date) < 2
    )
),
grouped AS (
    SELECT
        time,
        jsonb_object_agg(lesson::text, teachers ORDER BY lesson) AS lessons_map
    FROM filtered_slots
    GROUP BY time
)
SELECT jsonb_object_agg(time, lessons_map) AS result
FROM grouped
`

	var raw []byte
	if err = r.pool.QueryRow(ctx, query,
		filter.UserUUID,
		slotsJSON,
		filter.Type,
		filter.Topic,
		filter.Number,
	).Scan(&raw); err != nil {
		return nil, fmt.Errorf("booking repo: filter schedule: exec query: %w", err)
	}

	if raw == nil {
		return domain.Schedule{}, nil
	}

	result, err := convertJSONToSchedule(raw)
	if err != nil {
		return nil, fmt.Errorf("booking repo: filter schedule: convert json to schedule: %w", err)
	}

	return result, nil
}

func convertScheduleToJSON(schedule domain.Schedule) ([]byte, error) {
	return json.Marshal(schedule)
}

func convertJSONToSchedule(data []byte) (domain.Schedule, error) {
	var result domain.Schedule
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
