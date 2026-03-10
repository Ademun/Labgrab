package dto

type GetUserResDTO struct {
	Name             *string `json:"name"`
	Surname          *string `json:"surname"`
	Patronymic       *string `json:"patronymic"`
	GroupCode        *string `json:"group_code"`
	PhoneNumber      *string `json:"phone_number"`
	TelegramUsername string  `json:"telegram_username"`
	TelegramPhotoURL *string `json:"telegram_photo_url"`
	ApiReady         bool    `json:"api_ready"`
}
