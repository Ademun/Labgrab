package dikidi

import (
	"context"
	"encoding/json"
	"fmt"
	"labgrab/pkg/config"
	"net/url"
	"sync"
)

type Client struct {
	httpClient    *AdaptiveHTTPClient
	slotSourceIDs []int
	cfg           *config.DikidiClientConfig
}

func NewClient(cfg *config.DikidiClientConfig, httpClient *AdaptiveHTTPClient) *Client {
	return &Client{
		httpClient:    httpClient,
		slotSourceIDs: make([]int, 0),
		cfg:           cfg,
	}
}

func (c *Client) GetSlotStream(ctx context.Context) chan *SlotResult {
	results := make(chan *SlotResult)
	rate := make(chan struct{}, 50)

	go func() {
		defer close(results)

		wg := sync.WaitGroup{}

		for _, sourceID := range c.slotSourceIDs {
			rate <- struct{}{}
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					<-rate
				}()
				result, err := c.ProcessSlotSource(ctx, sourceID)
				if err != nil {
					select {
					case results <- &SlotResult{nil, err}:
					case <-ctx.Done():
						return
					}
				}
				select {
				case results <- &SlotResult{result, nil}:
				case <-ctx.Done():
					return
				}
			}()
		}
	}()

	return results
}

func (c *Client) ProcessSlotSource(ctx context.Context, slotSourceID int) (*APISlotData, error) {
	initialData, err := c.FetchSlotSource(ctx, slotSourceID, nil)
	if err != nil {
		return nil, err
	}
	initialData.Data.ServiceID = slotSourceID

	dates := initialData.Data.DatesTrue
	if len(dates) == 0 {
		return initialData, nil
	}

	for _, date := range dates[1:] {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:

		}

		newData, err := c.FetchSlotSource(ctx, slotSourceID, &date)
		if err != nil {
			return nil, err
		}

		for k, v := range newData.Data.Masters {
			initialData.Data.Masters[k] = v
		}
		for k, v := range newData.Data.Times {
			if existing, exists := initialData.Data.Times[k]; exists {
				initialData.Data.Times[k] = append(existing, v...)
			} else {
				initialData.Data.Times[k] = v
			}
		}
	}

	return initialData, nil
}

func (c *Client) FetchSlotSource(ctx context.Context, slotSourceID int, date *string) (*APISlotData, error) {
	u, err := url.Parse(c.cfg.SourcesConfig.SlotsSourceURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if date != nil {
		q.Set("date", *date)
	}
	q.Set("service_id[]", fmt.Sprintf("%d", slotSourceID))
	u.RawQuery = q.Encode()

	res, err := c.httpClient.Get(ctx, u.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data APISlotData
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
