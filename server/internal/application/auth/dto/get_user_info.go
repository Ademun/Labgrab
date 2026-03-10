package dto

import "time"

type GetUserInfoResDTO struct {
	PhoneNumber string     `json:"phone_number"`
	Password    string     `json:"password"`
	ApiAuthed   bool       `json:"api_authed"`
	LastAuth    *time.Time `json:"last_auth"`
}
