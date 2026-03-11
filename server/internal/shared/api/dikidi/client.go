package dikidi

import (
	"context"
	"labgrab/pkg/config"
	"sync"

	"golang.org/x/time/rate"
)

type Client struct {
	parser     *Parser
	serviceIDs []int
	mu         sync.RWMutex
	limiter    *rate.Limiter
	cfg        *config.DikidiClientConfig
}

func NewClient(cfg *config.DikidiClientConfig, parser *Parser) *Client {
	return &Client{
		parser:     parser,
		serviceIDs: make([]int, 0),
		limiter:    rate.NewLimiter(rate.Limit(cfg.ApiRateLimit), cfg.ApiRateBurst),
		cfg:        cfg,
	}
}

func (c *Client) limitCall(ctx context.Context, call func()) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}
	call()
	return nil
}
