package auth

type TelegramAuthData struct {
	Id        int
	FirstName string
	LastName  string
	Username  string
	PhotoURL  string
	AuthDate  int
	Hash      string
}
