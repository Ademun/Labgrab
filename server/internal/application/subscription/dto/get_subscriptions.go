package dto

import "time"

type GetSubscriptionsResDTO struct {
	UUID          string     `json:"uuid"`
	LabType       string     `json:"lab_type"`
	LabTopic      string     `json:"lab_topic"`
	LabNumber     int        `json:"lab_number"`
	LabAuditorium *int       `json:"lab_auditorium"`
	Status        string     `json:"status"`
	AutoEnroll    bool       `json:"auto_enroll"`
	AnyDate       bool       `json:"any_date"`
	CreatedAt     time.Time  `json:"created_at"`
	ClosedAt      *time.Time `json:"closed_at"`
}
