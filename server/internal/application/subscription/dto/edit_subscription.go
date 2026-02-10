package dto

type EditSubscriptionReqDTO struct {
	SubscriptionUUID string  `json:"subscription_uuid"`
	LabType          *string `json:"lab_type"`
	LabTopic         *string `json:"lab_topic"`
	LabNumber        *int    `json:"lab_number"`
	LabAuditorium    *int    `json:"lab_auditorium"`
	Status           *string `json:"status"`
	AutoEnroll       *bool   `json:"auto_enroll"`
	AnyDate          *bool   `json:"any_date"`
}

type EditSubscriptionResDTO struct {
	UUID string `json:"uuid"`
}
