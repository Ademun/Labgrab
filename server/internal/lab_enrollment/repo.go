package lab_enrollment

import (
	"context"
	repo_errors "labgrab/internal/shared/errors"

	"github.com/Masterminds/squirrel"
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
	query, args, err := r.sq.Insert("lab_enrollment.user_data").
		Columns("uuid", "dikidi_password", "password_dek").
		Values(data.UUID, data.DikidiPassword, data.PasswordDEK).
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
			Procedure: "CreateUser",
			Step:      "Query execution",
			Err:       err,
		}
	}

	return err
}
