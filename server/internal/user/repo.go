package user

import (
	"context"
	"errors"
	"fmt"
	"labgrab/internal/shared/apperr"

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

func (r *Repo) CreateUser(ctx context.Context, user *DBUser) (uuid.UUID, error) {
	userUUID := uuid.New()

	query, args, err := r.sq.Insert("user_service.users").
		Columns("uuid", "name", "surname", "telegram_id", "telegram_username", "telegram_photo_url").
		Values(userUUID, user.Name, user.Surname, user.TelegramID, user.TelegramUsername, user.TelegramPhotoUrl).
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("user repo: create user: build query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user repo: create user: exec query: %w", err)
	}

	return userUUID, nil
}

func (r *Repo) GetUser(ctx context.Context, userUUID uuid.UUID) (*DBUser, error) {
	query, args, err := r.sq.Select(
		"name",
		"surname",
		"patronymic",
		"group_code",
		"phone_number",
		"telegram_id",
		"telegram_username",
		"telegram_photo_url",
		"api_ready",
	).
		From("user_service.users").
		Where(squirrel.Eq{"uuid": userUUID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("user repo: get user: build query: %w", err)
	}

	var user DBUser
	if err = r.pool.QueryRow(ctx, query, args...).Scan(
		&user.Name,
		&user.Surname,
		&user.Patronymic,
		&user.GroupCode,
		&user.PhoneNumber,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.TelegramPhotoUrl,
		&user.ApiReady,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, fmt.Errorf("user repo: get user: scan row: %w", err)
	}

	return &user, nil
}

func (r *Repo) UpdateUser(ctx context.Context, user *DBUser) error {
	query, args, err := r.sq.Update("user_service.users").
		Set("name", user.Name).
		Set("surname", user.Surname).
		Set("patronymic", user.Patronymic).
		Set("group_code", user.GroupCode).
		Set("phone_number", user.PhoneNumber).
		Where(squirrel.Eq{"uuid": user.UUID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("user repo: update user: build query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("user repo: update user: exec query: %w", err)
	}

	return nil
}

func (r *Repo) DeleteUser(ctx context.Context, userUUID uuid.UUID, tx pgx.Tx) error {
	query, args, err := r.sq.Delete("user_service.users").
		Where(squirrel.Eq{"user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("user repo: delete user: build query: %w", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("user repo: delete user: exec query: %w", err)
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
		return false, fmt.Errorf("user repo: exists by telegram id: build query: %w", err)
	}

	var exists bool
	if err = r.pool.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("user repo: exists by telegram id: scan row: %w", err)
	}

	return exists, nil
}

func (r *Repo) GetUserUUIDByTelegramID(ctx context.Context, telegramID int) (uuid.UUID, error) {
	query, args, err := r.sq.Select("uuid").
		From("user_service.users").
		Where(squirrel.Eq{"telegram_id": telegramID}).
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("user repo: get user uuid by telegram id: build query: %w", err)
	}

	var userUUID uuid.UUID
	if err = r.pool.QueryRow(ctx, query, args...).Scan(&userUUID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperr.ErrNotFound
		}
		return uuid.Nil, fmt.Errorf("user repo: get user uuid by telegram id: scan row: %w", err)
	}

	return userUUID, nil
}
