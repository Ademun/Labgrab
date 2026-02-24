package lab_enrollment

import "github.com/google/uuid"

type DBUserData struct {
	UUID              uuid.UUID `db:"uuid"`
	DikidiPhoneNumber string    `db:"dikidi_phone_number"`
	DikidiPassword    string    `db:"dikidi_password"`
	PasswordDEK       string    `db:"password_dek"`
	SessionCookie     *string   `db:"password_cookie"`
	NoiseCookies      *string   `db:"noise_cookies"`
}
