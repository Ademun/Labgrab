package dto

type NewSubscriptionReqDTO struct {
	UserUUID      string `json:"user_uuid"`
	LabType       string `json:"lab_type"`
	LabTopic      string `json:"lab_topic"`
	LabNumber     int    `json:"lab_number"`
	LabAuditorium *int   `json:"lab_auditorium"`
	CreatedAt     int64  `json:"created_at"`
}

type NewSubscriptionResDTO struct {
	UUID string `json:"uuid"`
}
