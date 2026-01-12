package api_client

import (
	"encoding/json"
	"fmt"
)

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

type APISlotData struct {
	Data APIServiceData `json:"data"`
}

type APIServiceData struct {
	ServiceID int
	Masters   APIMasters `json:"masters"`
	DatesTrue []string   `json:"dates_true"`
	Times     APITimes   `json:"times"`
}

type APIMasters map[int]APIMasterData

func (m *APIMasters) UnmarshalJSON(b []byte) error {
	var emptySlice []interface{}
	if err := json.Unmarshal(b, &emptySlice); err == nil {
		*m = make(map[int]APIMasterData)
		return nil
	}

	var masterMap map[int]APIMasterData
	if err := json.Unmarshal(b, &masterMap); err == nil {
		*m = masterMap
		return nil
	}

	return fmt.Errorf("unknown masters format")
}

type APIMasterData struct {
	Username    string `json:"username"`
	ServiceName string `json:"service_name"`
}

type APITimes map[int][]string

func (t *APITimes) UnmarshalJSON(b []byte) error {
	var emptySlice []interface{}
	if err := json.Unmarshal(b, &emptySlice); err == nil {
		*t = make(map[int][]string)
		return nil
	}
	var timesMap map[int][]string
	if err := json.Unmarshal(b, &timesMap); err == nil {
		*t = timesMap
		return nil
	}
	return fmt.Errorf("unknown times format")
}
