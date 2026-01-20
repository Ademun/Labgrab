package auth

import "time"

type TelegramAuthData struct {
	Id        int
	FirstName string
	LastName  string
	Username  string
	PhotoURL  string
	AuthDate  time.Time
	Hash      string
}
