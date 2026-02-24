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
			"Sec-Fetch-User": "?1",
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
