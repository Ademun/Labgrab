package dto

type NewUserReqDTO struct {
	TelegramID  int    `json:"telegram_id"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic"`
	GroupCode   string `json:"group_code"`
	PhoneNumber string `json:"phone_number"`
}

type NewUserRespDTO struct {
	UserUUID string `json:"user_uuid"`
}
