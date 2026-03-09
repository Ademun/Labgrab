package user

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// DBUser user_service.users
type DBUser struct {
	UUID             uuid.UUID `db:"uuid"`
	Name             *string   `db:"name"`
	Surname          *string   `db:"surname"`
	Patronymic       *string   `db:"patronymic"`
	GroupCode        *string   `db:"group_code"`
	PhoneNumber      *int      `db:"phone_number"`
	TelegramID       int       `db:"telegram_id"`
	TelegramUsername string    `db:"telegram_username"`
	TelegramPhotoUrl *string   `db:"telegram_photo_url"`
	ApiReady         bool      `db:"api_ready"`
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
	Name             *string
	Surname          *string
	Patronymic       *string
	GroupCode        *string
	PhoneNumber      *int
	TelegramID       int
	TelegramUsername string
	TelegramPhotoUrl *string
	ApiReady         bool
}

type UpdateUserReq struct {
	UserUUID    uuid.UUID
	Name        *string
	Surname     *string
	Patronymic  *string
	GroupCode   *string
	PhoneNumber *int
}
