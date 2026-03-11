package dto

import "time"

type GetUserInfoResDTO struct {
	PhoneNumber string     `json:"phone_number"`
	ApiAuthed   bool       `json:"api_authed"`
	LastAuth    *time.Time `json:"last_auth"`
}
