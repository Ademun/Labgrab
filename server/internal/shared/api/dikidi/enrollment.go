package dikidi

import (
	"context"
	"fmt"
	"labgrab/internal/shared/apperr"
	"strconv"

	"github.com/imroc/req/v3"
)

func (c *Client) AcquireTimeReservation(ctx context.Context, client *req.Client, req *EventReservationRequest) (*EventReservationResponse, error) {
	var reservationData APITimeReservation
	_, err := client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"company_id":    "550001",
			"master_id":     strconv.Itoa(req.EventID),
			"services_id[]": strconv.Itoa(req.ServicesID),
			"time":          req.Time,
			"action_source": "direct_link",
			"session":       req.Session,
		}).
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest":   "empty",
			"Sec-Fetch-Mode":   "cors",
			"Sec-Fetch-Site":   "same-origin",
			"X-Requested-With": "XMLHttpRequest",
			"Referer":          fmt.Sprintf("https://dikidi.net/550001?p=2.pi-ssm-sd&o=7&s=%d&rl=0_undefined", req.ServicesID),
		}).
		SetSuccessResult(&reservationData).
		Get("https://dikidi.net/ru/ajax/newrecord/time_reservation/")
	if err != nil {
		return nil, &apperr.ExternalAPIError{
			Procedure: "AcquireTimeReservation",
			Step:      "Request",
			Err:       err,
		}
	}

	eventID, err := strconv.Atoi(reservationData.MasterID)
	if err != nil {
		return nil, &apperr.ExternalAPIError{
			Procedure: "AcquireTimeReservation",
			Step:      "ID parsing",
			Err:       err,
		}
	}

	return &EventReservationResponse{
		BookingID:      reservationData.RecordID,
		EventID:        eventID,
		DurationString: reservationData.DurationString,
	}, nil
}

func (c *Client) CheckEnrollment(ctx context.Context, client *req.Client, req *EnrollmentCheckRequest) error {
	referer := fmt.Sprintf(
		"https://dikidi.net/550001?p=3.pi-ssm-sd-cf&o=7&m=%d&s=%d&d=%s&r=%d&rl=0_%d&sdr=",
		req.MasterID, req.ServicesID, req.Time, req.RecordID, req.RecordID,
	)

	var checkResp APIEnrollmentCheck
	resp, err := client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"company":    "550001",
			"session":    req.Session,
			"social_key": "",
		}).
		SetFormData(map[string]string{
			"company":                  "550001",
			"type":                     "normal",
			"session":                  req.Session,
			"social_key":               "",
			"share_id":                 "0",
			"phone":                    req.Phone,
			"first_name":               req.FirstName,
			"last_name":                req.LastName,
			"comments":                 req.Comments,
			"promocode_appointment_id": "",
		}).
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest":   "empty",
			"Sec-Fetch-Mode":   "cors",
			"Sec-Fetch-Site":   "same-origin",
			"Origin":           "https://dikidi.net",
			"Referer":          referer,
			"X-Requested-With": "XMLHttpRequest",
		}).
		SetSuccessResult(&checkResp).
		Post("https://dikidi.net/ru/mobile/newrecord/check/")
	if err != nil {
		return &apperr.ExternalAPIError{
			Procedure: "CheckEnrollment",
			Step:      "Request",
			Err:       err,
		}
	}
	fmt.Printf("[CheckEnrollment] response body: %s\n", resp.String())

	if checkResp.Error != 0 {
		return &apperr.ExternalAPIError{
			Procedure: "CheckEnrollment",
			Step:      "Check error field",
			Err:       fmt.Errorf("api returned error code: %d", checkResp.Error),
		}
	}

	return nil
}

func (c *Client) GetReservationInfo(ctx context.Context, client *req.Client, req *ReservationInfoRequest) error {
	referer := fmt.Sprintf(
		"https://dikidi.net/550001?p=3.pi-ssm-sd-cf&o=7&m=%d&s=%d&d=%s&r=%d&rl=0_%d&sdr=",
		req.MasterID, req.ServicesID, req.Time, req.BookingID, req.BookingID,
	)

	var infoResp APIReservationInfo
	resp, err := client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"company_id":       "550001",
			"record_id_list[]": fmt.Sprintf("%d", req.BookingID),
			"session":          req.Session,
		}).
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest":   "empty",
			"Sec-Fetch-Mode":   "cors",
			"Sec-Fetch-Site":   "same-origin",
			"X-Requested-With": "XMLHttpRequest",
			"Referer":          referer,
		}).
		SetSuccessResult(&infoResp).
		Get("https://dikidi.net/ru/mobile/ajax/newrecord/records_info/")
	if err != nil {
		return &apperr.ExternalAPIError{
			Procedure: "GetReservationInfo",
			Step:      "Request",
			Err:       err,
		}
	}
	fmt.Printf("[GetReservationInfo] response body: %s\n", resp.String())

	if infoResp.Error.Code != 0 {
		msg := "<nil>"
		if infoResp.Error.Message != nil {
			msg = *infoResp.Error.Message
		}
		return &apperr.ExternalAPIError{
			Procedure: "GetReservationInfo",
			Step:      "Check error field",
			Err:       fmt.Errorf("api returned error code: %d, message: %s", infoResp.Error.Code, msg),
		}
	}

	return nil
}

func (c *Client) CreateBooking(ctx context.Context, client *req.Client, req *CreateBookingRequest) (int, error) {
	referer := fmt.Sprintf(
		"https://dikidi.net/550001?p=3.pi-ssm-sd-cf&o=7&m=%d&s=%d&d=%s&r=%d&rl=0_%d&sdr=",
		req.EventID, req.ServiceID, req.Time, req.BookingID, req.BookingID,
	)

	var createResp APICreateRecord
	resp, err := client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"company_id": "550001",
			"session":    req.Session,
			"social_key": "",
			"action":     "send_code_info_continue_1",
			"unique_num": "1",
		}).
		SetFormData(map[string]string{
			"type":              "normal",
			"name":              req.FirstName,
			"first_name":        req.FirstName,
			"last_name":         req.LastName,
			"phone":             req.Phone,
			"code":              "",
			"comments":          req.Comments,
			"is_show_all_times": "3",
			"captcha_token":     "",
			"action_source":     "direct link",
			"session":           req.Session,
			"social_key":        "",
			"active_cart_id":    "0",
			"active_method":     "0",
			"agreement":         "1",
		}).
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest":   "empty",
			"Sec-Fetch-Mode":   "cors",
			"Sec-Fetch-Site":   "same-origin",
			"X-Requested-With": "XMLHttpRequest",
			"Referer":          referer,
		}).
		SetSuccessResult(&createResp).
		Post("https://dikidi.net/ru/ajax/newrecord/record/")
	if err != nil {
		return 0, &apperr.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Request",
			Err:       err,
		}
	}
	fmt.Printf("[CreateRecord] response body: %s\n", resp.String())

	if len(createResp.Bookings) == 0 {
		return 0, &apperr.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Check booking",
			Err:       fmt.Errorf("empty booking in response"),
		}
	}

	if createResp.Bookings[0].Status != "1" {
		return 0, &apperr.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Check status",
			Err:       fmt.Errorf("unexpected booking status: %s", createResp.Bookings[0].Status),
		}
	}

	id, err := strconv.Atoi(createResp.Bookings[0].ID)
	if err != nil {
		return 0, &apperr.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Convert id",
			Err:       err,
		}
	}

	return id, nil
}
