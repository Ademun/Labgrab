package dto

type GetUserRespDTO struct {
	Username    string  `json:"username"`
	Name        *string `json:"name"`
	Surname     *string `json:"surname"`
	Patronymic  *string `json:"patronymic"`
	GroupCode   *string `json:"group_code"`
	PhoneNumber *int    `json:"phone_number"`
	PhotoURL    *string `json:"photo_url"`
	ApiReady    bool    `json:"api_ready"`
}
