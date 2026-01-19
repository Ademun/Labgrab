package user

import "github.com/google/uuid"

// DBUserDetails user_service.users_details
type DBUserDetails struct {
	Name       string    `db:"name"`
	Surname    string    `db:"surname"`
	Patronymic *string   `db:"patronymic"`
	GroupCode  string    `db:"group_code"`
	UserUUID   uuid.UUID `db:"user_uuid"`
}

// DBUserContacts user_service.users_contacts
type DBUserContacts struct {
	PhoneNumber string    `db:"phone_number"`
	Email       *string   `db:"email"`
	TelegramID  *int64    `db:"telegram_id"`
	UserUUID    uuid.UUID `db:"user_uuid"`
}

type DBUserInfo struct {
	UUID        uuid.UUID `db:"uuid"`
	Name        string    `db:"name"`
	Surname     string    `db:"surname"`
	Patronymic  *string   `db:"patronymic"`
	GroupCode   string    `db:"group_code"`
	PhoneNumber string    `db:"phone_number"`
	TelegramID  *int64    `db:"telegram_id"`
}

type CreateUserReq struct {
	Name        string
	Surname     string
	Patronymic  *string
	GroupCode   string
	PhoneNumber string
	Email       *string
	TelegramID  *int64
}

type CreateUserRes struct {
	UUID uuid.UUID
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
