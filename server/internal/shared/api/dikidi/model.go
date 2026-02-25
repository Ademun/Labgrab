package dikidi

type SlotResult struct {
	Data *APISlotData
	Err  error
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

type SlotReservationRequest struct {
	MasterID   int
	ServicesID int
	Time       string
	Session    string
}

type SlotReservationResponse struct {
	RecordID       int
	MasterID       string
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
