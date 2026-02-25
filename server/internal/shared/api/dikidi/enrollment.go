package dikidi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
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
			"Sec-Fetch-Dest": "empty",
			"Sec-Fetch-Mode": "cors",
			"Sec-Fetch-Site": "same-origin",
			"Origin":         "https://dikidi.net",
			"Referer":        fmt.Sprintf("https://dikidi.net/550001?p=2.pi-ssm-sd&o=7&s=%d&rl=0_undefined", req.ServicesID),
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

	var reservationData TimeReservationResponse
	if err := json.NewDecoder(reader).Decode(&reservationData); err != nil {
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
			"Sec-Fetch-Dest": "empty",
			"Sec-Fetch-Mode": "cors",
			"Sec-Fetch-Site": "same-site",
			"Origin":         "https://dikidi.net",
			"Referer":        referer,
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

	var checkResp EnrollmentCheckResponse
	if err := json.NewDecoder(reader).Decode(&checkResp); err != nil {
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
			"Sec-Fetch-Dest": "empty",
			"Sec-Fetch-Mode": "cors",
			"Sec-Fetch-Site": "same-site",
			"Origin":         "https://dikidi.net",
			"Referer":        referer,
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

	var infoResp ReservationInfoResponse
	if err := json.NewDecoder(reader).Decode(&infoResp); err != nil {
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
