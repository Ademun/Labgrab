package lab_enrollment

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DBUserData struct {
	UserUUID          uuid.UUID `db:"user_uuid"`
	DikidiPhoneNumber string    `db:"dikidi_phone_number"`
	DikidiPassword    string    `db:"dikidi_password"`
	PasswordDEK       string    `db:"password_dek"`
	Session           *string   `db:"session"`
	Token             *string   `db:"token"`
	NoiseCookies      *string   `db:"noise_cookies"`
}

type DBUserCookies struct {
	Session      *string `db:"session"`
	Token        *string `db:"token"`
	NoiseCookies *string `db:"noise_cookies"`
}

type CreateUserDataReq struct {
	UserUUID          uuid.UUID
	DikidiPhoneNumber string
	DikidiPassword    string
	Tx                pgx.Tx
}
