package event

import (
	"labgrab/internal/shared/domain"
	"time"

	"github.com/google/uuid"
)

type EnrollmentReq struct {
	UserUUID    uuid.UUID
	EventID     int
	ServiceID   int
	Time        time.Time
	Name        string
	Surname     string
	Patronymic  string
	Group       string
	PhoneNumber string
	Session     string
	Cookies     string
}

type GetEventsRes struct {
	Data *domain.Event
	Err  error
}
