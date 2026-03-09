package dikidi

import (
	"context"
	"fmt"

	"github.com/imroc/req/v3"
)

func (c *Client) GetBookings(ctx context.Context, client *req.Client, req *GetBookingsRequest) (*GetBookingsResult, error) {
	var apiResp APIGetRecords
	var err error
	c.limitCall(func() {
		_, err = client.R().
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

	})
	if err != nil {
		return nil, fmt.Errorf("api client: get bookings: request failed: %w", err)
	}

	if apiResp.Error.Code != 0 {
		msg := "<nil>"
		if apiResp.Error.Message != nil {
			msg = *apiResp.Error.Message
		}
		return nil, fmt.Errorf("api client: get bookings: bad error code, message: %s", msg)
	}

	active, err := c.parser.ParseRecords(apiResp.Data.New.List)
	if err != nil {
		return nil, fmt.Errorf("api client: get bookings: parsing new record: %w", err)
	}

	closed, err := c.parser.ParseRecords(apiResp.Data.Old.List)
	if err != nil {
		return nil, fmt.Errorf("api client: get bookings: parsing old records: %w", err)
	}

	return &GetBookingsResult{
		Active: active,
		Closed: closed,
	}, nil
}

func (c *Client) RemoveBooking(ctx context.Context, client *req.Client, req *RemoveBookingRequest) error {
	var removeResp APIRemoveRecord
	var err error
	c.limitCall(func() {
		_, err = client.R().
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
	})
	if err != nil {
		return fmt.Errorf("api client: remove bookings: request failed: %w", err)
	}

	if removeResp.Error != 0 {
		return fmt.Errorf("api client: remove bookings: bad error code: %d", removeResp.Error)
	}

	return nil
}
