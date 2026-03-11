package dikidi

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/imroc/req/v3"
)

func (c *Client) UpdateServiceIDs(ctx context.Context, client *req.Client) error {
	services, err := c.ScrapeServices(ctx, client, c.cfg.ServiceProviderURL)
	if err != nil {
		return fmt.Errorf("api client: update service ids: scraping: %w", err)
	}

	c.mu.Lock()
	c.serviceIDs = services
	c.mu.Unlock()

	return nil
}

func (c *Client) GetEventStream(ctx context.Context, client *req.Client) chan *GetEventsResult {
	c.mu.RLock()
	ids := make([]int, len(c.serviceIDs))
	copy(ids, c.serviceIDs)
	c.mu.RUnlock()

	results := make(chan *GetEventsResult)
	rate := make(chan struct{}, 3)

	go func() {
		wg := sync.WaitGroup{}

		for _, sourceID := range c.serviceIDs {
			rate <- struct{}{}
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					<-rate
				}()
				result, err := c.ProcessService(ctx, client, sourceID)
				if err != nil {
					select {
					case results <- &GetEventsResult{nil, err}:
					case <-ctx.Done():
						return
					}
					return
				}

				parsed, err := c.parser.ParseServiceData(&result.Data)
				if err != nil {
					select {
					case results <- &GetEventsResult{nil, err}:
					case <-ctx.Done():
						return
					}
					return
				}

				for _, event := range parsed {
					select {
					case results <- &GetEventsResult{
						Event: &event,
						Error: nil,
					}:
					case <-ctx.Done():
						return
					}
				}
			}()
		}

		wg.Wait()
		close(results)
	}()

	return results
}

func (c *Client) ProcessService(ctx context.Context, client *req.Client, serviceID int) (*APIService, error) {
	initialData, err := c.FetchService(ctx, client, serviceID, nil)
	if err != nil {
		return nil, fmt.Errorf("api client: process service: initial fetch: %w", err)
	}
	initialData.Data.ServiceID = serviceID

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

		newData, err := c.FetchService(ctx, client, serviceID, &date)
		if err != nil {
			return nil, fmt.Errorf("api client: process service: consecutive fetch: %w", err)
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

func (c *Client) FetchService(ctx context.Context, client *req.Client, serviceID int, date *string) (*APIService, error) {
	params := map[string]string{
		"day_month":    "",
		"service_id[]": strconv.Itoa(serviceID),
	}

	if date != nil {
		params["date"] = *date
		params["with_first"] = "false"
	}
	var data APIService
	var err error
	if err := c.limitCall(ctx, func() {
		_, err = client.R().
			SetContext(ctx).
			SetQueryParams(params).
			SetHeaders(map[string]string{
				"Sec-Fetch-Dest": "empty",
				"Sec-Fetch-Mode": "cors",
				"Sec-Fetch-Site": "same-site",
				"Origin":         "https://dikidi.net",
				"Referer":        "https://dikidi.net/550001?p=0.pi-ssm",
			}).
			SetSuccessResult(&data).
			Get(c.cfg.EventProviderURL)
	}); err != nil {
		return nil, fmt.Errorf("api client: fetch service: failed to acquire rate: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("api cliennt: fetch service: request failed: %w", err)
	}
	return &data, nil
}
