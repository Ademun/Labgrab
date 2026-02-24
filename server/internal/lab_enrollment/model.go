package lab_enrollment

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DBUserData struct {
	UUID              uuid.UUID `db:"uuid"`
	DikidiPhoneNumber string    `db:"dikidi_phone_number"`
	DikidiPassword    string    `db:"dikidi_password"`
	PasswordDEK       string    `db:"password_dek"`
	SessionCookie     *string   `db:"password_cookie"`
	NoiseCookies      *string   `db:"noise_cookies"`
}

type CreateuserDataReq struct {
	UserUUID          uuid.UUID
	DikidiPhoneNumber string
	DikidiPassword    string
	tx                pgx.Tx
}
