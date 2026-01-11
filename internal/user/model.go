package user

import "github.com/google/uuid"

type dbUserDetails struct {
	Name       string    `db:"name"`
	Surname    string    `db:"surname"`
	Patronymic *string   `db:"patronymic"`
	GroupCode  string    `db:"group_code"`
	UserUUID   uuid.UUID `db:"user_uuid"`
}

type dbUserContacts struct {
	PhoneNumber string    `db:"phone_number"`
	Email       *string   `db:"email"`
	TelegramID  *int64    `db:"telegram_id"`
	UserUUID    uuid.UUID `db:"user_uuid"`
}

type CreateUserRes struct {
	UUID uuid.UUID
}

type CreateUserDetailsReq struct {
	UserUUID   uuid.UUID
	Name       string
	Surname    string
	Patronymic *string
	GroupCode  string
}

type CreateUserContactsReq struct {
	UserUUID    uuid.UUID
	PhoneNumber string
	Email       *string
	TelegramID  *int64
}

type GetUserInfoRes struct {
	UUID        uuid.UUID
	Name        string
	Surname     string
	Patronymic  *string
	GroupCode   string
	PhoneNumber string
	TelegramID  *int64
}

type UpdateUserDetailsReq struct {
	UserUUID   uuid.UUID
	Name       string
	Surname    string
	Patronymic *string
	GroupCode  string
}

type UpdateUserContactsReq struct {
	UserUUID    uuid.UUID
	PhoneNumber string
	Email       *string
	TelegramID  *int64
}

type DBUser struct {
	UUID uuid.UUID `db:"uuid"`
}
