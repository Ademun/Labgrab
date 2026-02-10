package dto

type NewSubscriptionReqDTO struct {
	LabType       string `json:"lab_type"`
	LabTopic      string `json:"lab_topic"`
	LabNumber     int    `json:"lab_number"`
	LabAuditorium *int   `json:"lab_auditorium"`
	AutoEnroll    bool   `json:"auto_enroll"`
	AnyDate       bool   `json:"any_date"`
	CreatedAt     int64  `json:"created_at"`
}

type NewSubscriptionResDTO struct {
	UUID string `json:"uuid"`
}
