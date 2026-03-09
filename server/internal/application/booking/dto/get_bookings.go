package dto

import (
	"labgrab/internal/booking"
	"labgrab/internal/shared/domain"
	"time"
)

type GetBookingsRespDTO struct {
	Type       domain.LabType  `json:"type"`
	Topic      domain.LabTopic `json:"topic"`
	Number     int             `json:"number"`
	Auditorium int             `json:"auditorium"`
	Spot       *int            `json:"spot"`
	Lesson     domain.Lesson   `json:"lesson"`
	Start      time.Time       `json:"start_time"`
	End        time.Time       `json:"end_time"`
	Status     booking.Status  `json:"status"`
}
