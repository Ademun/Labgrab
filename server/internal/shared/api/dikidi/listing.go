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

func (c *Client) GetRecords(ctx context.Context, client *req.Client, req *GetRecordsRequest) (*GetRecordsResult, error) {
	resp := client.Get("https://dikidi.net/ru/mobile/ajax/newrecord/get_records/").
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
		Do(ctx)
	if resp.Err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Request",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, &errors.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Request",
			Err:       fmt.Errorf("bad status code: %d", resp.StatusCode),
		}
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Read from body",
			Err:       err,
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Read from body",
			Err:       err,
		}
	}
	fmt.Printf("[GetRecords] response body: %s\n", body)

	var apiResp GetRecordsResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "GetRecords",
			Step:      "Parse from body",
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

	return &GetRecordsResult{
		New: mapRecords(apiResp.Data.New.List),
		Old: mapRecords(apiResp.Data.Old.List),
	}, nil
}

func mapRecords(list []APIRecord) []Record {
	records := make([]Record, 0, len(list))
	for _, r := range list {
		rec := Record{
			ID:       r.ID,
			Time:     r.Time,
			TimeTo:   r.TimeTo,
			Duration: r.Duration,
		}
		if len(r.Services) > 0 {
			rec.ServiceName = r.Services[0].Name
		}
		if len(r.Employees) > 0 {
			rec.MasterName = r.Employees[0].Username
		}
		records = append(records, rec)
	}
	return records
}
