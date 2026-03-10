package dto

type UpdateUserReqDTO struct {
	Name        *string `json:"name"`
	Surname     *string `json:"surname"`
	Patronymic  *string `json:"patronymic"`
	GroupCode   *string `json:"group_code"`
	PhoneNumber *string `json:"phone_number"`
}
