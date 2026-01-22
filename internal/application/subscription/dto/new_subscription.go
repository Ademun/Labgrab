package dto

import "time"

type NewSubscriptionReqDTO struct {
	UserUUID      string    `json:"user_uuid"`
	LabType       string    `json:"lab_type"`
	LabTopic      string    `json:"lab_topic"`
	LabNumber     int       `json:"lab_number"`
	LabAuditorium *int      `json:"lab_auditorium"`
	CreatedAt     time.Time `json:"created_at"`
}

type NewSubscriptionResDTO struct {
	UUID string `json:"uuid"`
}
