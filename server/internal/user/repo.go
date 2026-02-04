package user

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

func (r *Repo) CreateUser(ctx context.Context, user *DBUser, tx pgx.Tx) (uuid.UUID, error) {
	userUUID := uuid.New()
	query, args, err := r.sq.Insert("user_service.users").
		Columns("uuid", "name", "surname", "telegram_id", "username", "photo_url").
		Values(userUUID, user.Name, user.Surname, user.TelegramID, user.Username, user.PhotoUrl).
		ToSql()
	if err != nil {
		return userUUID, &repo_errors.ErrDBProcedure{
			Procedure: "CreateUser",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return userUUID, &repo_errors.ErrDBProcedure{
			Procedure: "CreateUser",
			Step:      "Query execution",
			Err:       err,
		}
	}

	return userUUID, err
}

func (r *Repo) GetUser(ctx context.Context, userUUID uuid.UUID) (*DBUser, error) {
	query, args, err := r.sq.Select(
		"username",
		"name",
		"surname",
		"patronymic",
		"group_code",
		"phone_number",
		"photo_url",
	).
		From("user_service.users").
		Where(squirrel.Eq{"uuid": userUUID}).
		ToSql()
	if err != nil {
		return nil, &repo_errors.ErrDBProcedure{
			Procedure: "GetUserInfo",
			Step:      "Query setup",
			Err:       err,
		}
	}
	var userInfo DBUser
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&userInfo.Username,
		&userInfo.Name,
		&userInfo.Surname,
		&userInfo.Patronymic,
		&userInfo.GroupCode,
		&userInfo.PhoneNumber,
		&userInfo.PhotoUrl,
	)

	if err != nil {
		return nil, &repo_errors.ErrDBProcedure{
			Procedure: "GetUserInfo",
			Step:      "Row scanning",
			Err:       err,
		}
	}

	return &userInfo, nil
}

func (r *Repo) UpdateUser(ctx context.Context, user *DBUser) error {
	query, args, err := r.sq.Update("user_service.users").
		Set("name", user.Name).
		Set("surname", user.Surname).
		Set("patronymic", user.Patronymic).
		Set("group_code", user.GroupCode).
		Set("phone_number", user.PhoneNumber).
		Set("photo_url", user.PhotoUrl).
		Where(squirrel.Eq{"uuid": user.UUID}).
		ToSql()
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "UpdateUserDetails",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "UpdateUserDetails",
			Step:      "Query execution",
			Err:       err,
		}
	}
	return nil
}

func (r *Repo) ExistsByTelegramID(ctx context.Context, telegramID int) (bool, error) {
	subquery := r.sq.Select("1").
		From("user_service.users").
		Where(squirrel.Eq{"telegram_id": telegramID}).
		Limit(1)

	query, args, err := r.sq.Select().
		Column(squirrel.Expr("EXISTS(?)", subquery)).
		ToSql()
	if err != nil {
		return false, &repo_errors.ErrDBProcedure{
			Procedure: "ExistsByTelegramID",
			Step:      "Query setup",
			Err:       err,
		}
	}

	var exists bool
	err = r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, &repo_errors.ErrDBProcedure{
			Procedure: "ExistsByTelegramID",
			Step:      "Row scanning",
			Err:       err,
		}
	}

	return exists, nil
}

func (r *Repo) GetUserUUIDByTelegramID(ctx context.Context, telegramID int) (uuid.UUID, error) {
	query, args, err := r.sq.Select("uuid").
		From("user_service.users").
		Where(squirrel.Eq{"telegram_id": telegramID}).
		ToSql()
	if err != nil {
		return uuid.Nil, &repo_errors.ErrDBProcedure{
			Procedure: "GetUserUUIDByTelegramID",
			Step:      "Query setup",
			Err:       err,
		}
	}

	var userUUID uuid.UUID
	err = r.pool.QueryRow(ctx, query, args...).Scan(&userUUID)
	if err != nil {
		return uuid.Nil, &repo_errors.ErrDBProcedure{
			Procedure: "GetUserUUIDByTelegramID",
			Step:      "Row scanning",
			Err:       err,
		}
	}

	return userUUID, nil
}
