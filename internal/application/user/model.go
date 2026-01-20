package user

type CreateUserReq struct {
	Name        string
	Surname     string
	Patronymic  *string
	GroupCode   string
	PhoneNumber string
	Email       *string
	TelegramID  *int64
}
