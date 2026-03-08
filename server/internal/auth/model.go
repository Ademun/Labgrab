package auth

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TelegramAuthData struct {
	Id        int
	FirstName string
	LastName  string
	Username  string
	PhotoURL  string
	AuthDate  int
	Hash      string
}

type DBUserData struct {
	UserUUID          uuid.UUID `db:"user_uuid"`
	DikidiPhoneNumber string    `db:"dikidi_phone_number"`
	DikidiPassword    string    `db:"dikidi_password"`
	DEK               string    `db:"dek"`
	Session           *string   `db:"session"`
	Token             *string   `db:"token"`
	Cookies           *string   `db:"cookies"`
}

type DBUserCookies struct {
	Session *string `db:"session"`
	Token   *string `db:"token"`
	Cookies *string `db:"cookies"`
}

type DecryptedUserData struct {
	UserUUID          uuid.UUID
	DikidiPhoneNumber string
	DikidiPassword    string
	Session           *string
	Token             *string
	Cookies           *string
}

type CreateUserDataReq struct {
	UserUUID          uuid.UUID
	DikidiPhoneNumber string
	DikidiPassword    string
	Tx                pgx.Tx
}

type GetUserCredentialsRes struct {
	DikidiPhoneNumber string
	Session           *string
	Token             *string
	Cookies           *string
}

type ErrHashIntegrity struct {
	ExpectedHash string
	ActualHash   string
}

func (e ErrHashIntegrity) Error() string {
	return fmt.Sprintf("hash integrity check failed: expected=%s actual=%s", e.ExpectedHash, e.ActualHash)
}

type ErrAuthDateExpired struct {
	AuthDate    time.Time
	CurrentDate time.Time
}

func (e ErrAuthDateExpired) Error() string {
	return fmt.Sprintf("auth date expired: diff=%.2fh (limit=24h)", e.CurrentDate.Sub(e.AuthDate).Hours())
}
