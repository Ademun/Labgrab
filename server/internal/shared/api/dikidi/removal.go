package dikidi

import (
	"context"
	"fmt"
	"labgrab/internal/shared/errors"

	"github.com/imroc/req/v3"
)

func (c *Client) RemoveRecord(ctx context.Context, client *req.Client, req *RemoveRecordRequest) error {
	var removeResp RemoveRecordResponse
	resp, err := client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"id":         req.RecordID,
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
		return &errors.ExternalAPIError{
			Procedure: "RemoveRecord",
			Step:      "Request",
			Err:       err,
		}
	}
	fmt.Printf("[RemoveRecord] response body: %s\n", resp.String())

	if removeResp.Error != 0 {
		return &errors.ExternalAPIError{
			Procedure: "RemoveRecord",
			Step:      "Check error field",
			Err:       fmt.Errorf("api returned error code: %d", removeResp.Error),
		}
	}

	return nil
}
