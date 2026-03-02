package dikidi

import (
	"encoding/json"
	"fmt"
)

type APIHTMLPageOptions struct {
	StepData APIHTMLStepData `json:"step_data"`
}

type APIHTMLStepData struct {
	List []APIHTMLList `json:"list"`
}

type APIHTMLList struct {
	Services []APIHTMLService `json:"services"`
}

type APIHTMLService struct {
	ID int `json:"id"`
}

type APIService struct {
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

type APIAuth struct {
	HTML string `json:"html"`
}

type APITimeReservation struct {
	RecordID       int    `json:"record_id"`
	MasterID       string `json:"master_id"`
	DurationString string `json:"duration_string"`
}

type APIEnrollmentCheck struct {
	Error int `json:"error"`
}

type APIReservationInfo struct {
	Error APIReservationInfoError `json:"error"`
}

type APIReservationInfoError struct {
	Code    int     `json:"code"`
	Message *string `json:"message"`
}

type APICreateRecord struct {
	Bookings []APIBookingRecord `json:"bookings"`
}

type APIBookingRecord struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type APIGetRecords struct {
	Error APIGetRecordsError `json:"error"`
	Data  APIGetRecordsData  `json:"data"`
}

type APIGetRecordsError struct {
	Code    int     `json:"code"`
	Message *string `json:"message"`
}

type APIGetRecordsData struct {
	New APIGetRecordsList `json:"new"`
	Old APIGetRecordsList `json:"old"`
}

type APIGetRecordsList struct {
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

type APIRemoveRecord struct {
	Error int `json:"error"`
}
