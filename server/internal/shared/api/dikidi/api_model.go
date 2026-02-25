package dikidi

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

type AuthResponse struct {
	HTML string `json:"html"`
}

type TimeReservationResponse struct {
	RecordID       int    `json:"record_id"`
	MasterID       string `json:"master_id"`
	DurationString string `json:"duration_string"`
}

type EnrollmentCheckResponse struct {
	Error int `json:"error"`
}

type ReservationInfoResponse struct {
	Error ReservationInfoError `json:"error"`
}

type ReservationInfoError struct {
	Code    int     `json:"code"`
	Message *string `json:"message"`
}

type CreateRecordResponse struct {
	Bookings []BookingRecord `json:"bookings"`
}

type BookingRecord struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type GetRecordsResponse struct {
	Error GetRecordsError `json:"error"`
	Data  GetRecordsData  `json:"data"`
}

type GetRecordsError struct {
	Code    int     `json:"code"`
	Message *string `json:"message"`
}

type GetRecordsData struct {
	New GetRecordsList `json:"new"`
	Old GetRecordsList `json:"old"`
}

type GetRecordsList struct {
	More bool        `json:"more"`
	List []APIRecord `json:"list"`
}

type APIRecord struct {
	ID        string              `json:"id"`
	Time      string              `json:"time"`
	TimeTo    string              `json:"time_to"`
	Duration  string              `json:"duration"`
	Services  []APIRecordService  `json:"services"`
	Employees []APIRecordEmployee `json:"employees"`
}

type APIRecordService struct {
	Name string `json:"name"`
}

type APIRecordEmployee struct {
	Username string `json:"username"`
}

type RemoveRecordResponse struct {
	Error int `json:"error"`
}
