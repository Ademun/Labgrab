package dto

type EditSubscriptionReqDTO struct {
	UserUUID         string  `json:"user_uuid"`
	SubscriptionUUID string  `json:"subscription_uuid"`
	LabType          *string `json:"lab_type"`
	LabTopic         *string `json:"lab_topic"`
	LabNumber        *int    `json:"lab_number"`
	LabAuditorium    *int    `json:"lab_auditorium"`
}

type EditSubscriptionResDTO struct {
	UUID string `json:"uuid"`
}
