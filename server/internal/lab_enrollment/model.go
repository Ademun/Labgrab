package lab_enrollment

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

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

type EnrollReq struct {
	UserUUID   uuid.UUID
	MasterID   int
	ServiceID  int
	Time       time.Time
	Name       string
	Surname    string
	Patronymic string
	Group      string
}

type GetRecordsReq struct {
	UserUUID uuid.UUID
}

type GetRecordsRes struct {
	New []RecordItem
	Old []RecordItem
}

type RecordItem struct {
	ID          string
	Time        string
	TimeTo      string
	Duration    string
	ServiceName string
	MasterName  string
}

type RemoveRecordReq struct {
	UserUUID uuid.UUID
	RecordID string
}
