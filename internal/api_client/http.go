package api_client

import (
	"context"
	"io"
	"labgrab/pkg/config"
	"math"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/time/rate"
)

type AdaptiveHTTPClient struct {
	client  *http.Client
	limiter *rate.Limiter
	cfg     *config.HTTPClientConfig
	mu      sync.Mutex
}

func NewAdaptiveHTTPClient(cfg *config.HTTPClientConfig) *AdaptiveHTTPClient {
	return &AdaptiveHTTPClient{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		limiter: rate.NewLimiter(cfg.MinRate, cfg.BurstSize),
		cfg:     cfg,
		mu:      sync.Mutex{},
	}
}

func (c *AdaptiveHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	res, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	c.handleHttpResponse(res)
	return res, nil
}

func (c *AdaptiveHTTPClient) Head(ctx context.Context, url string) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	res, err := c.client.Head(url)
	if err != nil {
		return nil, err
	}
	c.handleHttpResponse(res)
	return res, nil
}

func (c *AdaptiveHTTPClient) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	res, err := c.client.Post(url, contentType, body)
	if err != nil {
		return nil, err
	}
	c.handleHttpResponse(res)
	return res, nil
}

func (c *AdaptiveHTTPClient) PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	res, err := c.client.PostForm(url, data)
	if err != nil {
		return nil, err
	}
	c.handleHttpResponse(res)
	return res, nil
}

func (c *AdaptiveHTTPClient) handleHttpResponse(res *http.Response) {
	if res == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if res.StatusCode >= 400 {
		c.reduceRate()
	} else {
		c.increaseRate()
	}
}

func (c *AdaptiveHTTPClient) reduceRate() {
	newRate := math.Max(float64(c.limiter.Limit())*c.cfg.DecreaseFactor, float64(c.cfg.MinRate))
	c.limiter.SetLimit(rate.Limit(newRate))
}

func (c *AdaptiveHTTPClient) increaseRate() {
	newRate := math.Min(float64(c.limiter.Limit())*c.cfg.IncreaseFactor, float64(c.cfg.MaxRate))
	c.limiter.SetLimit(rate.Limit(newRate))
}
