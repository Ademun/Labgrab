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
	PhoneNumber      *string   `db:"phone_number"`
	TelegramID       int64     `db:"telegram_id"`
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
}

type GetUserRes struct {
	Name             *string
	Surname          *string
	Patronymic       *string
	GroupCode        *string
	PhoneNumber      *string
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
	PhoneNumber *string
}

type DeleteUserReq struct {
	UserUUID uuid.UUID
	Tx       pgx.Tx
}
