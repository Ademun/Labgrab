package event

import (
	"context"
	repo_errors "labgrab/internal/shared/errors"

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

func (r *Repo) CreateUserData(ctx context.Context, data *DBUserData, tx pgx.Tx) error {
	query, args, err := r.sq.Insert("lab_enrollment_service.user_data").
		Columns("user_uuid", "dikidi_phone_number", "dikidi_password", "dek").
		Values(data.UserUUID, data.DikidiPhoneNumber, data.DikidiPassword, data.DEK).
		ToSql()
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "CreateUserData",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "CreateUserData",
			Step:      "Query execution",
			Err:       err,
		}
	}

	return nil
}

func (r *Repo) GetUserData(ctx context.Context, userUUID uuid.UUID) (*DBUserData, error) {
	query, args, err := r.sq.Select("user_uuid", "dikidi_phone_number", "dikidi_password", "dek", "session", "token", "cookies").
		From("lab_enrollment_service.user_data").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return nil, &repo_errors.ErrDBProcedure{
			Procedure: "GetUserData",
			Step:      "Query setup",
			Err:       err,
		}
	}

	var data DBUserData
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&data.UserUUID,
		&data.DikidiPhoneNumber,
		&data.DikidiPassword,
		&data.DEK,
		&data.Session,
		&data.Token,
		&data.Cookies,
	)
	if err != nil {
		return nil, &repo_errors.ErrDBProcedure{
			Procedure: "GetUserData",
			Step:      "Row scanning",
			Err:       err,
		}
	}

	return &data, nil
}

func (r *Repo) SetUserCookies(ctx context.Context, userUUID uuid.UUID, cookies *DBUserCookies) error {
	query, args, err := r.sq.Update("lab_enrollment_service.user_data").
		Set("session", cookies.Session).
		Set("token", cookies.Token).
		Set("cookies", cookies.Cookies).
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "SetUserCookies",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "SetUserCookies",
			Step:      "Query execution",
			Err:       err,
		}
	}

	return nil
}
