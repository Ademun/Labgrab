package user

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repo struct {
	pool *pgxpool.Conn
	sq   squirrel.StatementBuilderType
}

func NewRepo(pool *pgxpool.Conn) *Repo {
	return &Repo{pool: pool, sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *Repo) CreateUser(ctx context.Context) (uuid.UUID, error) {
	userUUID := uuid.New()
	query, args, err := r.sq.Insert("user_service.users").Columns("uuid").Values(userUUID).ToSql()
	if err != nil {
		return userUUID, err
	}

	_, err = r.pool.Exec(ctx, query, args...)

	return userUUID, err
}

func (r *Repo) CreateUserDetails(ctx context.Context, details *DBUserDetails) error {
	query, args, err := r.sq.Insert("user_service.users_details").
		Columns("name", "surname", "patronymic", "group_code", "user_uuid").
		Values(details.Name, details.Surname, details.Patronymic, details.GroupCode, details.UserUUID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, query, args...)
	return err
}

func (r *Repo) CreateUserContacts(ctx context.Context, contacts *DBUserContacts) error {
	query, args, err := r.sq.Insert("user_service.users_contacts").
		Columns("phone_number", "email", "telegram_id", "user_uuid").
		Values(contacts.PhoneNumber, contacts.Email, contacts.TelegramID, contacts.UserUUID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, query, args...)
	return err
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
		return nil, err
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
		return nil, err
	}

	return &userInfo, nil
}

func (r *Repo) UpdateUserDetails(ctx context.Context, req *UpdateUserDetailsReq) error {
	query, args, err := r.sq.Update("user_service.users_details").
		Set("name", req.Name).
		Set("surname", req.Surname).
		Set("patronymic", req.Patronymic).
		Set("group_code", req.GroupCode).
		Where(squirrel.Eq{"user_uuid": req.UserUUID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	return err
}

func (r *Repo) UpdateUserContacts(ctx context.Context, req *UpdateUserContactsReq) error {
	query, args, err := r.sq.Update("user_service.users_contacts").
		Set("phone_number", req.PhoneNumber).
		Set("email", req.Email).
		Set("telegram_id", req.TelegramID).
		Where(squirrel.Eq{"user_uuid": req.UserUUID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	return err
}
