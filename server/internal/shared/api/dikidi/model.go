package dikidi

import (
	"labgrab/internal/shared/domain"
)

type GetEventsResult struct {
	Event *domain.Event
	Error error
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
	EventID        int
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
	BookingID  int
	MasterID   int
	ServicesID int
	Time       string // YYYYMMDDHHmm
	Session    string
}

type CreateBookingRequest struct {
	EventID   int
	ServiceID int
	Time      string // YYYYMMDDHHmm
	BookingID int
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
	Active []domain.Booking
	Closed []domain.Booking
}

type RemoveBookingRequest struct {
	BookingID string
	Session   string
}
