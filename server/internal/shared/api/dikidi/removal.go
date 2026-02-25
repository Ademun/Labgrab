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

func (c *Client) RemoveRecord(ctx context.Context, client *req.Client, req *RemoveRecordRequest) error {
	resp := client.Get("https://dikidi.net/ru/mobile/newrecord/remove_record/").
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
		Do(ctx)
	if resp.Err != nil {
		return &errors.ExternalAPIError{
			Procedure: "RemoveRecord",
			Step:      "Request",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &errors.ExternalAPIError{
			Procedure: "RemoveRecord",
			Step:      "Request",
			Err:       fmt.Errorf("bad status code: %d", resp.StatusCode),
		}
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return &errors.ExternalAPIError{
			Procedure: "RemoveRecord",
			Step:      "Read from body",
			Err:       err,
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return &errors.ExternalAPIError{
			Procedure: "RemoveRecord",
			Step:      "Read from body",
			Err:       err,
		}
	}
	fmt.Printf("[RemoveRecord] response body: %s\n", body)

	var removeResp RemoveRecordResponse
	if err := json.Unmarshal(body, &removeResp); err != nil {
		return &errors.ExternalAPIError{
			Procedure: "RemoveRecord",
			Step:      "Parse from body",
			Err:       err,
		}
	}

	if removeResp.Error != 0 {
		return &errors.ExternalAPIError{
			Procedure: "RemoveRecord",
			Step:      "Check error field",
			Err:       fmt.Errorf("api returned error code: %d", removeResp.Error),
		}
	}

	return nil
}
