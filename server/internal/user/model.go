package user

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// DBUser user_service.users
type DBUser struct {
	UUID        uuid.UUID `db:"uuid"`
	Name        *string   `db:"name"`
	Surname     *string   `db:"surname"`
	Patronymic  *string   `db:"patronymic"`
	GroupCode   *string   `db:"group_code"`
	PhoneNumber *int      `db:"phone_number"`
	TelegramID  int       `db:"telegram_id"`
	Username    string    `db:"username"`
	PhotoUrl    *string   `db:"photo_url"`
}

type CreateUserReq struct {
	TelegramID int
	Name       *string
	Surname    *string
	Username   string
	PhotoUrl   *string
	Tx         pgx.Tx
}

type CreateUserRes struct {
	UUID uuid.UUID
	Tx   pgx.Tx
}

type GetUserRes struct {
	Username    string
	Name        *string
	Surname     *string
	Patronymic  *string
	GroupCode   *string
	PhoneNumber *int
	TelegramID  int
	PhotoUrl    *string
}

type UpdateUserReq struct {
	UserUUID    uuid.UUID
	Name        *string
	Surname     *string
	Patronymic  *string
	GroupCode   *string
	PhoneNumber *int
	PhotoUrl    *string
}
