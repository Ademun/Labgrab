package dikidi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"labgrab/internal/shared/errors"

	"github.com/imroc/req/v3"
)

func (c *Client) AcquireTimeReservation(ctx context.Context, client *req.Client, req *SlotReservationRequest) (*SlotReservationResponse, error) {
	resp := client.Get("https://dikidi.net/ru/ajax/newrecord/time_reservation/").
		SetQueryParams(map[string]string{
			"company_id":    "550001",
			"master_id":     fmt.Sprintf("%d", req.MasterID),
			"services_id[]": fmt.Sprintf("%d", req.ServicesID),
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
		}).Do(ctx)
	if resp.Err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "AcquireTimeReservation",
			Step:      "Request",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, &errors.ExternalAPIError{
			Procedure: "AcquireTimeReservation",
			Step:      "Request",
			Err:       fmt.Errorf("bad status code: %d", resp.StatusCode),
		}
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "AcquireTimeReservation",
			Step:      "Read from body",
			Err:       err,
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "AcquireTimeReservation",
			Step:      "Read from body",
			Err:       err,
		}
	}
	fmt.Printf("[AcquireTimeReservation] response body: %s\n", body)

	var reservationData TimeReservationResponse
	if err := json.Unmarshal(body, &reservationData); err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "AcquireTimeReservation",
			Step:      "Parse from body",
			Err:       err,
		}
	}

	return &SlotReservationResponse{
		RecordID:       reservationData.RecordID,
		MasterID:       reservationData.MasterID,
		DurationString: reservationData.DurationString,
	}, nil
}

func (c *Client) CheckEnrollment(ctx context.Context, client *req.Client, req *EnrollmentCheckRequest) error {
	referer := fmt.Sprintf(
		"https://dikidi.net/550001?p=3.pi-ssm-sd-cf&o=7&m=%d&s=%d&d=%s&r=%d&rl=0_%d&sdr=",
		req.MasterID, req.ServicesID, req.Time, req.RecordID, req.RecordID,
	)

	resp := client.Post("https://dikidi.net/ru/mobile/newrecord/check/").
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
		Do(ctx)
	if resp.Err != nil {
		return &errors.ExternalAPIError{
			Procedure: "CheckEnrollment",
			Step:      "Request",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &errors.ExternalAPIError{
			Procedure: "CheckEnrollment",
			Step:      "Request",
			Err:       fmt.Errorf("bad status code: %d", resp.StatusCode),
		}
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return &errors.ExternalAPIError{
			Procedure: "CheckEnrollment",
			Step:      "Read from body",
			Err:       err,
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return &errors.ExternalAPIError{
			Procedure: "CheckEnrollment",
			Step:      "Read from body",
			Err:       err,
		}
	}
	fmt.Printf("[CheckEnrollment] response body: %s\n", body)

	var checkResp EnrollmentCheckResponse
	if err := json.Unmarshal(body, &checkResp); err != nil {
		return &errors.ExternalAPIError{
			Procedure: "CheckEnrollment",
			Step:      "Parse from body",
			Err:       err,
		}
	}

	if checkResp.Error != 0 {
		return &errors.ExternalAPIError{
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
		req.MasterID, req.ServicesID, req.Time, req.RecordID, req.RecordID,
	)

	resp := client.Get("https://dikidi.net/ru/mobile/ajax/newrecord/records_info/").
		SetQueryParams(map[string]string{
			"company_id":       "550001",
			"record_id_list[]": fmt.Sprintf("%d", req.RecordID),
			"session":          req.Session,
		}).
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest":   "empty",
			"Sec-Fetch-Mode":   "cors",
			"Sec-Fetch-Site":   "same-origin",
			"X-Requested-With": "XMLHttpRequest",
			"Referer":          referer,
		}).
		Do(ctx)
	if resp.Err != nil {
		return &errors.ExternalAPIError{
			Procedure: "GetReservationInfo",
			Step:      "Request",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &errors.ExternalAPIError{
			Procedure: "GetReservationInfo",
			Step:      "Request",
			Err:       fmt.Errorf("bad status code: %d", resp.StatusCode),
		}
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return &errors.ExternalAPIError{
			Procedure: "GetReservationInfo",
			Step:      "Read from body",
			Err:       err,
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return &errors.ExternalAPIError{
			Procedure: "GetReservationInfo",
			Step:      "Read from body",
			Err:       err,
		}
	}
	fmt.Printf("[GetReservationInfo] response body: %s\n", body)

	var infoResp ReservationInfoResponse
	if err := json.Unmarshal(body, &infoResp); err != nil {
		return &errors.ExternalAPIError{
			Procedure: "GetReservationInfo",
			Step:      "Parse from body",
			Err:       err,
		}
	}

	if infoResp.Error.Code != 0 {
		msg := "<nil>"
		if infoResp.Error.Message != nil {
			msg = *infoResp.Error.Message
		}
		return &errors.ExternalAPIError{
			Procedure: "GetReservationInfo",
			Step:      "Check error field",
			Err:       fmt.Errorf("api returned error code: %d, message: %s", infoResp.Error.Code, msg),
		}
	}

	return nil
}

func (c *Client) CreateRecord(ctx context.Context, client *req.Client, req *CreateRecordRequest) (*CreateRecordResponse, error) {
	referer := fmt.Sprintf(
		"https://dikidi.net/550001?p=3.pi-ssm-sd-cf&o=7&m=%d&s=%d&d=%s&r=%d&rl=0_%d&sdr=",
		req.MasterID, req.ServicesID, req.Time, req.RecordID, req.RecordID,
	)

	resp := client.Post("https://dikidi.net/ru/ajax/newrecord/record/").
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
		Do(ctx)
	if resp.Err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Request",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, &errors.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Request",
			Err:       fmt.Errorf("bad status code: %d", resp.StatusCode),
		}
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Read from body",
			Err:       err,
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Read from body",
			Err:       err,
		}
	}
	fmt.Printf("[CreateRecord] response body: %s\n", body)

	var createResp CreateRecordResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Parse from body",
			Err:       err,
		}
	}

	if len(createResp.Bookings) == 0 {
		return nil, &errors.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Check bookings",
			Err:       fmt.Errorf("empty bookings in response"),
		}
	}

	if createResp.Bookings[0].Status != "1" {
		return nil, &errors.ExternalAPIError{
			Procedure: "CreateRecord",
			Step:      "Check status",
			Err:       fmt.Errorf("unexpected booking status: %s", createResp.Bookings[0].Status),
		}
	}

	return &createResp, nil
}
