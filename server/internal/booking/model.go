package booking

import (
	"labgrab/internal/shared/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Status string

const (
	StatusOpen   Status = "Open"
	StatusClosed Status = "Closed"
)

type DBBooking struct {
	BookingID  int             `db:"booking_id"`
	Type       domain.LabType  `db:"type"`
	Topic      domain.LabTopic `db:"topic"`
	Number     int             `db:"number"`
	Auditorium int             `db:"auditorium"`
	Spot       *int            `db:"spot"`
	Lesson     domain.Lesson   `db:"lesson"`
	Start      time.Time       `db:"start_time"`
	End        time.Time       `db:"end_time"`
	Status     Status          `db:"status"`
	UserUUID   uuid.UUID       `db:"user_uuid"`
}

type DBSlotFilter struct {
	UserUUID uuid.UUID
	Schedule domain.Schedule
	Type     domain.LabType
	Topic    domain.LabTopic
	Number   int
}

type GetBookingsRes struct {
	UserUUID uuid.UUID
	Status   Status
	Data     *domain.Booking
}

type LoadClientBookingsReq struct {
	UserUUID uuid.UUID
	Session  string
	Cookies  string
}

type CancelClientBookingReq struct {
	UserUUID  uuid.UUID
	BookingID int
	Session   string
	Cookies   string
}

type DeleteBookingsReq struct {
	UserUUID uuid.UUID
	Tx       pgx.Tx
}
type FilterScheduleReq struct {
	UserUUID uuid.UUID
	Schedule domain.Schedule
	Type     domain.LabType
	Topic    domain.LabTopic
	Number   int
}
