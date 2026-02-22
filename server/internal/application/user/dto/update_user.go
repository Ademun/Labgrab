package dto

type UpdateUserReqDTO struct {
	Name        *string `json:"name"`
	Surname     *string `json:"surname"`
	Patronymic  *string `json:"patronymic"`
	GroupCode   *string `json:"group_code"`
	PhoneNumber *int    `json:"phone_number"`
	PhotoUrl    *string `json:"photo_url"`
}
