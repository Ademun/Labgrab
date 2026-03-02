package dikidi

import (
	"labgrab/internal/shared/domain"
	"time"
)

type Service struct {
	ID int
}

type Event struct {
	ID         int
	ServiceID  int
	Name       string
	Type       domain.LabType
	Topic      domain.LabTopic
	Number     int
	Auditorium int
	Spot       *int
	Schedule   domain.Schedule
	Link       string
}

type Booking struct {
	ID         string
	Name       string
	Type       domain.LabType
	Topic      domain.LabTopic
	Number     int
	Auditorium int
	Spot       *int
	Start      time.Time
	End        time.Time
}

type CSRFTokenRequest struct {
	PhoneNumber       string
	TelegramCSRFToken string
}

type AuthRequest struct {
	PhoneNumber       string
	Password          string
	CSRFToken         string
	TelegramCSRFToken string
}

type ClientCookies struct {
	CookieName *string
	Token      *string
	All        string
}

type EventReservationRequest struct {
	EventID    int
	ServicesID int
	Time       string
	Session    string
}

type EventReservationResponse struct {
	EventID        string
	BookingID      int
	DurationString string
}

type EnrollmentCheckRequest struct {
	MasterID   int
	ServicesID int
	Time       string // YYYYMMDDHHmm
	RecordID   int
	Session    string
	Phone      string
	FirstName  string
	LastName   string
	Comments   string
}

type ReservationInfoRequest struct {
	RecordID   int
	MasterID   int
	ServicesID int
	Time       string // YYYYMMDDHHmm
	Session    string
}

type CreateBookingRequest struct {
	EventID   int
	ServiceID int
	Time      string // YYYYMMDDHHmm
	RecordID  int
	Session   string
	Phone     string
	FirstName string
	LastName  string
	Comments  string
}

type GetBookingsRequest struct {
	Session string
}

type GetBookingsResult struct {
	Active []Booking
	Closed []Booking
}

type RemoveBookingRequest struct {
	BookingID string
	Session   string
}
