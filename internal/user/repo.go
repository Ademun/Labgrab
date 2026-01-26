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

func (r *Repo) CreateUser(ctx context.Context, details *DBUserDetails, contacts *DBUserContacts) (uuid.UUID, pgx.Tx, error) {
	userUUID := uuid.New()
	query, args, err := r.sq.Insert("user_service.users").Columns("uuid").Values(userUUID).ToSql()
	if err != nil {
		return userUUID, nil, &repo_errors.ErrDBProcedure{
			Procedure: "CreateUser",
			Step:      "Query setup",
			Err:       err,
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return userUUID, nil, &repo_errors.ErrDBProcedure{
			Procedure: "CreateUser",
			Step:      "Transaction setup",
			Err:       err,
		}
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		tx.Rollback(ctx)
		return userUUID, nil, &repo_errors.ErrDBProcedure{
			Procedure: "CreateUser",
			Step:      "Query execution",
			Err:       err,
		}
	}

	details.UserUUID = userUUID
	contacts.UserUUID = userUUID

	if err = r.createUserDetails(ctx, details, tx); err != nil {
		tx.Rollback(ctx)
		return userUUID, nil, err
	}

	if err = r.createUserContacts(ctx, contacts, tx); err != nil {
		tx.Rollback(ctx)
		return userUUID, nil, err
	}

	return userUUID, tx, err
}

func (r *Repo) createUserDetails(ctx context.Context, details *DBUserDetails, tx pgx.Tx) error {
	query, args, err := r.sq.Insert("user_service.users_details").
		Columns("name", "surname", "patronymic", "group_code", "user_uuid").
		Values(details.Name, details.Surname, details.Patronymic, details.GroupCode, details.UserUUID).
		ToSql()
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "CreateUserDetails",
			Step:      "Query setup",
			Err:       err,
		}
	}
	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "CreateUserDetails",
			Step:      "Query execution",
			Err:       err,
		}
	}
	return nil
}

func (r *Repo) createUserContacts(ctx context.Context, contacts *DBUserContacts, tx pgx.Tx) error {
	query, args, err := r.sq.Insert("user_service.users_contacts").
		Columns("phone_number", "telegram_id", "user_uuid").
		Values(contacts.PhoneNumber, contacts.TelegramID, contacts.UserUUID).
		ToSql()
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "CreateUserContacts",
			Step:      "Query setup",
			Err:       err,
		}
	}
	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "CreateUserContacts",
			Step:      "Query execution",
			Err:       err,
		}
	}
	return nil
}

func (r *Repo) GetUserInfo(ctx context.Context, userUUID uuid.UUID) (*DBUserInfo, error) {
	query, args, err := r.sq.Select(
		"ud.user_uuid AS uuid",
		"ud.name",
		"ud.surname",
		"ud.patronymic",
		"ud.group_code",
		"uc.phone_number",
		"uc.telegram_id",
	).
		From("user_service.users_details AS ud").
		InnerJoin("user_service.users_contacts AS uc ON ud.user_uuid = uc.user_uuid").
		Where(squirrel.Eq{"ud.user_uuid": userUUID}).
		ToSql()
	if err != nil {
		return nil, &repo_errors.ErrDBProcedure{
			Procedure: "GetUserInfo",
			Step:      "Query setup",
			Err:       err,
		}
	}
	var userInfo DBUserInfo
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&userInfo.UUID,
		&userInfo.Name,
		&userInfo.Surname,
		&userInfo.Patronymic,
		&userInfo.GroupCode,
		&userInfo.PhoneNumber,
		&userInfo.TelegramID,
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

func (r *Repo) UpdateUserDetails(ctx context.Context, details *DBUserDetails) error {
	query, args, err := r.sq.Update("user_service.users_details").
		Set("name", details.Name).
		Set("surname", details.Surname).
		Set("patronymic", details.Patronymic).
		Set("group_code", details.GroupCode).
		Where(squirrel.Eq{"user_uuid": details.UserUUID}).
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

func (r *Repo) UpdateUserContacts(ctx context.Context, contacts *DBUserContacts) error {
	query, args, err := r.sq.Update("user_service.users_contacts").
		Set("phone_number", contacts.PhoneNumber).
		Set("telegram_id", contacts.TelegramID).
		Where(squirrel.Eq{"user_uuid": contacts.UserUUID}).
		ToSql()
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "UpdateUserContacts",
			Step:      "Query setup",
			Err:       err,
		}
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return &repo_errors.ErrDBProcedure{
			Procedure: "UpdateUserContacts",
			Step:      "Query execution",
			Err:       err,
		}
	}
	return nil
}

func (r *Repo) ExistsByTelegramID(ctx context.Context, telegramID int) (bool, error) {
	subquery := r.sq.Select("1").
		From("user_service.users_contacts").
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
