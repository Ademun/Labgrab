package user

type CreateUserReqDTO struct {
	Name        string  `json:"name"`
	Surname     string  `json:"surname"`
	Patronymic  *string `json:"patronymic"`
	GroupCode   string  `json:"group_code"`
	PhoneNumber string  `json:"phone_number"`
	Email       *string `json:"email"`
	TelegramID  *int64  `json:"telegram_id"`
}
