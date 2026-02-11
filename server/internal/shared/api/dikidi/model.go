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
	SessionID  *string
	Other      map[string]string
}
