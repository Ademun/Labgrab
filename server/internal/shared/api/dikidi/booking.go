package dikidi

import (
	"context"
	"fmt"
	"labgrab/internal/shared/apperr"

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
		return nil, &apperr.ExternalAPIError{
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
		return nil, &apperr.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Check error field",
			Err:       fmt.Errorf("api returned error code: %d, message: %s", apiResp.Error.Code, msg),
		}
	}

	active, err := c.parser.ParseRecords(apiResp.Data.New.List)
	if err != nil {
		return nil, &apperr.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Map records",
			Err:       err,
		}
	}

	closed, err := c.parser.ParseRecords(apiResp.Data.Old.List)
	if err != nil {
		return nil, &apperr.ExternalAPIError{
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

func (c *Client) RemoveBooking(ctx context.Context, client *req.Client, req *RemoveBookingRequest) error {
	var removeResp APIRemoveRecord
	resp, err := client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"id":         req.BookingID,
			"session":    req.Session,
			"social_key": "",
		}).
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest":   "empty",
			"Sec-Fetch-Mode":   "cors",
			"Sec-Fetch-Site":   "same-origin",
			"X-Requested-With": "XMLHttpRequest",
			"Referer":          "https://dikidi.net/550001?p=1.pi-ur",
		}).
		SetSuccessResult(&removeResp).
		Get("https://dikidi.net/ru/mobile/newrecord/remove_record/")
	if err != nil {
		return &apperr.ExternalAPIError{
			Procedure: "RemoveRecord",
			Step:      "Request",
			Err:       err,
		}
	}
	fmt.Printf("[RemoveRecord] response body: %s\n", resp.String())

	if removeResp.Error != 0 {
		return &apperr.ExternalAPIError{
			Procedure: "RemoveRecord",
			Step:      "Check error field",
			Err:       fmt.Errorf("api returned error code: %d", removeResp.Error),
		}
	}

	return nil
}
