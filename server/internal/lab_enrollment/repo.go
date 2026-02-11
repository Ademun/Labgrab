package lab_enrollment

import (
	"context"
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

func (r *Repo) CreateJobs(ctx context.Context, jobs []*DBEnrollmentJob) error {
	jobUUID := uuid.New()

	builder := r.sq.Insert("lab_enrollment_service.jobs").
		Columns(
			"job_uuid",
			"user_uuid",
			"subscription_uuid",
			"job_status",
			"available_dates",
			"started_at",
			"completed_at",
		)
	for _, job := range jobs {
		builder = builder.Values(jobUUID, job.UserUUID, job.SubscriptionUUID, job.Status, job.AvailableDates, job.StartedAt, job.CompletedAt)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CreateJobs",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "CreateJobs",
			Step:      "Query execution",
			Err:       err,
		}
	}

	return nil
}

func (r *Repo) DeleteJobInfo(ctx context.Context, jobUUID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "DeleteJobInfo",
			Step:      "Begin transaction",
			Err:       err,
		}
	}

	query, args, err := r.sq.Delete("lab_enrollment_service.jobs").
		Where(squirrel.Eq{"job_uuid": jobUUID}).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "DeleteJobInfo",
			Step:      "Query setup",
			Err:       err,
		}
	}
	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "DeleteJobInfo",
			Step:      "Query execution",
			Err:       err,
		}
	}

	query, args, err = r.sq.Delete("lab_enrollment_service.job_results").
		Where(squirrel.Eq{"job_uuid": jobUUID}).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "DeleteJobInfo",
			Step:      "Query setup",
			Err:       err,
		}
	}
	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "DeleteJobInfo",
			Step:      "Query execution",
			Err:       err,
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "DeleteJobInfo",
			Step:      "Commit transaction",
			Err:       err,
		}
	}
	return nil
}

func (r *Repo) RecordJobSuccess(ctx context.Context, jobUUID uuid.UUID, enrollment *DBEnrollment) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobSuccess",
			Step:      "Begin transaction",
			Err:       err,
		}
	}

	updateJobQuery, updateJobArgs, err := r.sq.Update("lab_enrollment_service.jobs").
		Set("status", JobStatusCompleted).
		Set("completed_at", time.Now()).
		Where(squirrel.Eq{"job_uuid": jobUUID}).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobSuccess",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, updateJobQuery, updateJobArgs...)
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobSuccess",
			Step:      "Query execution",
			Err:       err,
		}
	}

	jobResultQuery, jobResultArgs, err := r.sq.Insert("lab_enrollment_service.job_results").
		Columns("job_uuid", "result", "error_message", "enrollment_uuid").
		Values(jobUUID, JobResultSuccess, nil, enrollment.UUID).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobSuccess",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, jobResultQuery, jobResultArgs...)
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobSuccess",
			Step:      "Query execution",
			Err:       err,
		}
	}

	enrollmentQuery, enrollmentArgs, err := r.sq.Insert("lab_enrollment_service.enrollments").
		Columns("enrollment_uuid", "user_uuid", "dikidi_enrollment_id", "visit_time", "enrolled_at").
		Values(enrollment.UUID, enrollment.UserUUID, enrollment.DikidiEnrollmentID, enrollment.VisitTime, enrollment.EnrolledAt).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobSuccess",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, enrollmentQuery, enrollmentArgs...)
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobSuccess",
			Step:      "Query execution",
			Err:       err,
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobSuccess",
			Step:      "Commit transaction",
			Err:       err,
		}
	}

	return nil
}

func (r *Repo) RecordJobFailure(ctx context.Context, jobUUID uuid.UUID, errorMessage string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobFailure",
			Step:      "Begin transaction",
			Err:       err,
		}
	}

	updateJobQuery, updateJobArgs, err := r.sq.Update("lab_enrollment_service.jobs").
		Set("status", JobStatusCompleted).
		Set("completed_at", time.Now()).
		Where(squirrel.Eq{"job_uuid": jobUUID}).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobFailure",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, updateJobQuery, updateJobArgs...)
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobFailure",
			Step:      "Query execution",
			Err:       err,
		}
	}

	jobResultQuery, jobResultArgs, err := r.sq.Insert("lab_enrollment_service.job_results").
		Columns("job_uuid", "result", "error_message", "enrollment_uuid").
		Values(jobUUID, JobResultFailed, errorMessage, nil).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobFailure",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, jobResultQuery, jobResultArgs...)
	if err != nil {
		tx.Rollback(ctx)
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobFailure",
			Step:      "Query execution",
			Err:       err,
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return &errors.ErrDBProcedure{
			Procedure: "RecordJobFailure",
			Step:      "Commit transaction",
			Err:       err,
		}
	}

	return nil
}

func (r *Repo) AcquireJob(ctx context.Context) (*DBEnrollmentJob, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "AcquireJob",
			Step:      "Begin transaction",
			Err:       err,
		}
	}

	query := `
			SELECT job_uuid, user_uuid, subscription_uuid, status, available_dates, started_at, completed_at
			FROM lab_enrollment_service.jobs
			WHERE status = $1
			ORDER BY started_at
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		`

	var job DBEnrollmentJob
	err = tx.QueryRow(ctx, query, JobStatusQueued).Scan(
		&job.UUID,
		&job.UserUUID,
		&job.SubscriptionUUID,
		&job.Status,
		&job.AvailableDates,
		&job.StartedAt,
		&job.CompletedAt,
	)
	if err != nil {
		tx.Rollback(ctx)
		return nil, &errors.ErrDBProcedure{
			Procedure: "AcquireJob",
			Step:      "Query execution",
			Err:       err,
		}
	}

	updateQuery, updateArgs, err := r.sq.Update("lab_enrollment_service.jobs").
		Set("status", JobStatusProcessing).
		Where(squirrel.Eq{"job_uuid": job.UUID}).
		ToSql()
	if err != nil {
		tx.Rollback(ctx)
		return nil, &errors.ErrDBProcedure{
			Procedure: "AcquireJob",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, updateQuery, updateArgs...)
	if err != nil {
		tx.Rollback(ctx)
		return nil, &errors.ErrDBProcedure{
			Procedure: "AcquireJob",
			Step:      "Query execution",
			Err:       err,
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, &errors.ErrDBProcedure{
			Procedure: "AcquireJob",
			Step:      "Commit transaction",
			Err:       err,
		}
	}

	job.Status = JobStatusProcessing

	return &job, nil
}
