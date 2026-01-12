package api_client

type HTMLPageOptions struct {
	StepData HTMLStepData `json:"step_data"`
}

type HTMLStepData struct {
	List []HTMLList `json:"list"`
}

type HTMLList struct {
	Services []HTMLService `json:"services"`
}

type HTMLService struct {
	ID int `json:"id"`
}
