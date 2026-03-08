package auth

import (
	"context"
	"fmt"

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
	return &Repo{
		pool: pool,
		sq:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *Repo) CreateUserData(ctx context.Context, data *DBUserData, tx pgx.Tx) error {
	query, args, err := r.sq.Insert("auth_service.user_data").
		Columns("user_uuid", "dikidi_phone_number", "dikidi_password", "dek").
		Values(data.UserUUID, data.DikidiPhoneNumber, data.DikidiPassword, data.DEK).
		ToSql()
	if err != nil {
		return fmt.Errorf("auth repo: create user data: build query: %w", err)
	}

	if _, err = tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("auth repo: create user data: exec query: %w", err)
	}

	return nil
}

func (r *Repo) GetUserData(ctx context.Context, userUUID uuid.UUID) (*DBUserData, error) {
	query, args, err := r.sq.Select(
		"user_uuid", "dikidi_phone_number", "dikidi_password", "dek", "session", "token", "cookies",
	).
		From("auth_service.user_data").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("auth repo: get user data: build query: %w", err)
	}

	var data DBUserData
	if err = r.pool.QueryRow(ctx, query, args...).Scan(
		&data.UserUUID,
		&data.DikidiPhoneNumber,
		&data.DikidiPassword,
		&data.DEK,
		&data.Session,
		&data.Token,
		&data.Cookies,
	); err != nil {
		return nil, fmt.Errorf("auth repo: get user data: scan row: %w", err)
	}

	return &data, nil
}

func (r *Repo) SetUserCookies(ctx context.Context, userUUID uuid.UUID, cookies *DBUserCookies) error {
	query, args, err := r.sq.Update("auth_service.user_data").
		Set("session", cookies.Session).
		Set("token", cookies.Token).
		Set("cookies", cookies.Cookies).
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("auth repo: set user cookies: build query: %w", err)
	}

	if _, err = r.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("auth repo: set user cookies: exec query: %w", err)
	}

	return nil
}
