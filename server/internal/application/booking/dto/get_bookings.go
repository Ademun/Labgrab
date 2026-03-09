package dto

import "time"

type GetBookingsResDTO struct {
	Type       string    `json:"type"`
	Topic      string    `json:"topic"`
	Number     int       `json:"number"`
	Auditorium int       `json:"auditorium"`
	Spot       *int      `json:"spot"`
	Lesson     int       `json:"lesson"`
	Start      time.Time `json:"start_time"`
	End        time.Time `json:"end_time"`
	Status     string    `json:"status"`
}
