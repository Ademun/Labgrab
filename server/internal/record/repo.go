package record

import (
	"context"
	"encoding/json"
	"fmt"
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
	sq   squirrel.StatementBuilderType
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool, sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func cronJobName(recordID int) string {
	return fmt.Sprintf("close-record-%d", recordID)
}

func endTimeToCronExpr(t time.Time) string {
	ceil := t.Add(time.Minute).Truncate(time.Minute)
	return fmt.Sprintf("%d %d %d %d *", ceil.Minute(), ceil.Hour(), ceil.Day(), int(ceil.Month()))
}

func (r *Repo) CreateRecord(ctx context.Context, rec *DBRecord) error {
	query, args, err := r.sq.Insert("record_service.records").
		Columns("record_id", "lab_type", "lab_topic", "lab_auditorium", "lesson", "start_time", "end_time", "status", "user_uuid").
		Values(rec.RecordID, rec.LabType, rec.LabTopic, rec.LabAuditorium, rec.Lesson, rec.StartTime, rec.EndTime, StatusOpen, rec.UserUUID).
		ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CreateRecord",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CreateRecord",
			Step:      "Query execution",
			Err:       err,
		}
	}

	if err = r.scheduleCronClose(ctx, rec.RecordID, rec.EndTime); err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CreateRecord",
			Step:      "pg_cron scheduling",
			Err:       err,
		}
	}

	return nil
}

func (r *Repo) scheduleCronClose(ctx context.Context, recordID int, endTime time.Time) error {
	jobName := cronJobName(recordID)
	cronExpr := endTimeToCronExpr(endTime)

	// SAFE: recordID is int, jobName derives solely from recordID — no user input reaches here.
	jobSQL := fmt.Sprintf(
		`UPDATE record_service.records SET status = 'Closed' WHERE record_id = %d AND status = 'Open' AND end_time <= NOW(); `+
			`SELECT cron.unschedule('%s');`,
		recordID, jobName,
	)

	_, err := r.pool.Exec(ctx,
		`SELECT cron.schedule($1, $2, $3)`,
		jobName, cronExpr, jobSQL,
	)
	return err
}

func (r *Repo) GetRecords(ctx context.Context, userUUID uuid.UUID) ([]DBRecord, error) {
	query, args, err := r.sq.Select(
		"record_id",
		"lab_type",
		"lab_topic",
		"lab_auditorium",
		"lesson",
		"start_time",
		"end_time",
		"status",
		"user_uuid",
	).
		From("record_service.records").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetRecords",
			Step:      "Query setup",
			Err:       err,
		}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetRecords",
			Step:      "Query execution",
			Err:       err,
		}
	}
	defer rows.Close()

	var records []DBRecord
	for rows.Next() {
		var rec DBRecord
		err = rows.Scan(
			&rec.RecordID,
			&rec.LabType,
			&rec.LabTopic,
			&rec.LabAuditorium,
			&rec.Lesson,
			&rec.StartTime,
			&rec.EndTime,
			&rec.Status,
			&rec.UserUUID,
		)
		if err != nil {
			return nil, &errors.ErrDBProcedure{
				Procedure: "GetRecords",
				Step:      "Row scanning",
				Err:       err,
			}
		}
		records = append(records, rec)
	}

	if err = rows.Err(); err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "GetRecords",
			Step:      "Row error check",
			Err:       err,
		}
	}

	return records, nil
}

func (r *Repo) DeleteRecord(ctx context.Context, recordID int) error {
	query, args, err := r.sq.Delete("record_service.records").
		Where(squirrel.Eq{"record_id": recordID}).
		ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "DeleteRecord",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "DeleteRecord",
			Step:      "Query execution",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx,
		`SELECT cron.unschedule(jobid) FROM cron.job WHERE jobname = $1`,
		cronJobName(recordID),
	)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "DeleteRecord",
			Step:      "pg_cron unschedule",
			Err:       err,
		}
	}

	return nil
}

func (r *Repo) CloseExpiredRecords(ctx context.Context) error {
	query, args, err := r.sq.Update("record_service.records").
		Set("status", StatusClosed).
		Where(squirrel.And{
			squirrel.Expr("end_time <= NOW()"),
			squirrel.Eq{"status": StatusOpen},
		}).
		ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CloseExpiredRecords",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CloseExpiredRecords",
			Step:      "Query execution",
			Err:       err,
		}
	}

	return nil
}

func (r *Repo) FilterAvailableSlots(ctx context.Context, filter *DBSlotFilter) (domain.Schedule, error) {
	slotsJSON, err := convertAvailableSlotsToJSON(filter.MatchingTimeslots)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "FilterAvailableSlots",
			Step:      "JSON conversion",
			Err:       err,
		}
	}

	query := `
WITH input_slots AS (
    SELECT
        times.key  AS time,
        lessons.key::int AS lesson,
        lessons.value    AS teachers
    FROM jsonb_each($2::jsonb) AS times,
         LATERAL jsonb_each(times.value) AS lessons
),
user_lessons AS (
    SELECT
        lesson,
        start_time::date AS lesson_date
    FROM record_service.records
    WHERE user_uuid = $1
      AND status = 'Open'
),
filtered_slots AS (
    SELECT
        i.time,
        i.lesson,
        i.teachers
    FROM input_slots i
    WHERE NOT EXISTS (
        SELECT 1
        FROM user_lessons ul
        WHERE ul.lesson      = i.lesson
          AND ul.lesson_date = i.time::timestamptz::date
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
	err = r.pool.QueryRow(ctx, query, filter.UserUUID, slotsJSON).Scan(&raw)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "FilterAvailableSlots",
			Step:      "Query execution",
			Err:       err,
		}
	}

	if raw == nil {
		return domain.Schedule{}, nil
	}

	result, err := convertJSONToMatchingTimeslots(raw)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "FilterAvailableSlots",
			Step:      "JSON conversion",
			Err:       err,
		}
	}

	return result, nil
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
