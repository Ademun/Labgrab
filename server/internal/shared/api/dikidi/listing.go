package dikidi

import (
	"context"
	"fmt"
	"labgrab/internal/shared/errors"
	"time"

	"github.com/imroc/req/v3"
)

func (c *Client) GetBookings(ctx context.Context, client *req.Client, req *GetBookingsRequest) (*GetBookingsResult, error) {
	var apiResp APIGetRecords
	_, err := client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"company":    "550001",
			"session":    req.Session,
			"social_key": "",
			"fresh":      "both",
		}).
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest":   "empty",
			"Sec-Fetch-Mode":   "cors",
			"Sec-Fetch-Site":   "same-origin",
			"X-Requested-With": "XMLHttpRequest",
			"Referer":          "https://dikidi.net/550001?p=0.pi-ssm",
		}).
		SetSuccessResult(&apiResp).
		Get("https://dikidi.net/ru/mobile/ajax/newrecord/get_records/")
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Request",
			Err:       err,
		}
	}

	if apiResp.Error.Code != 0 {
		msg := "<nil>"
		if apiResp.Error.Message != nil {
			msg = *apiResp.Error.Message
		}
		return nil, &errors.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Check error field",
			Err:       fmt.Errorf("api returned error code: %d, message: %s", apiResp.Error.Code, msg),
		}
	}

	active, err := mapBookings(apiResp.Data.New.List)
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Map records",
			Err:       err,
		}
	}

	closed, err := mapBookings(apiResp.Data.Old.List)
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Map records",
			Err:       err,
		}
	}

	return &GetBookingsResult{
		Active: active,
		Closed: closed,
	}, nil
}

func mapBookings(list []APIRecord) ([]Booking, error) {
	bookings := make([]Booking, 0, len(list))
	for _, r := range list {
		startTime, err := time.Parse(r.Time, "2006-01-02 15:04:05")
		if err != nil {
			return nil, err
		}
		endTime, err := time.Parse(r.Time, "2006-01-02 15:04:05")
		rec := Booking{
			ID:    r.ID,
			Start: startTime,
			End:   endTime,
		}
		bookings = append(bookings, rec)
	}
	return bookings, nil
}
