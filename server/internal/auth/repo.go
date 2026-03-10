package auth

import (
	"context"
	"errors"
	"fmt"
	"labgrab/internal/shared/apperr"
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
	return &Repo{
		pool: pool,
		sq:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *Repo) CreateUserData(ctx context.Context, data *DBUserData) error {
	query, args, err := r.sq.Insert("auth_service.user_data").
		Columns("user_uuid", "dikidi_phone_number", "dikidi_password", "dek").
		Values(data.UserUUID, data.DikidiPhoneNumber, data.DikidiPassword, data.DEK).
		ToSql()
	if err != nil {
		return fmt.Errorf("auth repo: create user data: build query: %w", err)
	}

	if _, err = r.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("auth repo: create user data: exec query: %w", err)
	}

	return nil
}

func (r *Repo) GetUserData(ctx context.Context, userUUID uuid.UUID) (*DBUserData, error) {
	query, args, err := r.sq.Select(
		"user_uuid", "dikidi_phone_number", "dikidi_password", "dek", "session", "token", "cookies", "api_authed", "last_auth",
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
		&data.ApiAuthed,
		&data.LastAuth,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, fmt.Errorf("auth repo: get user data: scan row: %w", err)
	}

	return &data, nil
}

func (r *Repo) UpdateUserData(ctx context.Context, data *DBUserData) error {
	query, args, err := r.sq.Update("auth_service.user_data").
		Set("dikidi_phone_number", data.DikidiPhoneNumber).
		Set("dikidi_password", data.DikidiPassword).
		Set("dek", data.DEK).
		Where(squirrel.Eq{"user_uuid": data.UserUUID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("auth repo: update user data: build query: %w", err)
	}

	if _, err = r.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("auth repo: update user data: exec query: %w", err)
	}

	return nil
}

func (r *Repo) DeleteUserData(ctx context.Context, userUUID uuid.UUID, tx pgx.Tx) error {
	query, args, err := r.sq.Delete("auth_service.user_data").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("auth repo: delete user data: build query: %w", err)
	}

	if _, err = tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("auth repo: delete user data: exec query: %w", err)
	}

	return nil
}

func (r *Repo) SetUserCookies(ctx context.Context, userUUID uuid.UUID, cookies *DBUserCookies) error {
	query, args, err := r.sq.Update("auth_service.user_data").
		Set("session", cookies.Session).
		Set("token", cookies.Token).
		Set("cookies", cookies.Cookies).
		Set("last_auth", time.Now().UTC()).
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

func (r *Repo) GetStaleUsers(ctx context.Context) ([]DBUserData, error) {
	query := `SELECT 
    user_uuid, 
    dikidi_phone_number,
    dikidi_password, 
    dek, 
    session, 
    token, 
    cookies, 
    api_authed, 
    last_auth FROM auth_service.user_data
	WHERE abs(timezone('UTC', now())::timestamp::date - last_auth::timestamp::date) > 20 OR api_authed IS NULL
`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("auth repo: get stale users: exec query: %w", err)
	}
	defer rows.Close()

	var users []DBUserData
	for rows.Next() {
		var user DBUserData
		if err := rows.Scan(
			&user.UserUUID,
			&user.DikidiPhoneNumber,
			&user.DikidiPassword,
			&user.DEK,
			&user.Session,
			&user.Token,
			&user.Cookies,
			&user.ApiAuthed,
			&user.LastAuth,
		); err != nil {
			return nil, fmt.Errorf("auth repo: get stale users: scan row: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("auth repo: get stale users: rows error: %w", err)
	}

	return users, nil
}
