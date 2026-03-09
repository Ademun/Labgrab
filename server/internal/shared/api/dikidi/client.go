package dikidi

import (
	"labgrab/pkg/config"
)

type Client struct {
	parser     *Parser
	serviceIDs []int
	limiter    chan struct{}
	cfg        *config.DikidiClientConfig
}

func NewClient(cfg *config.DikidiClientConfig, parser *Parser) *Client {
	return &Client{
		parser:     parser,
		serviceIDs: make([]int, 0),
		limiter:    make(chan struct{}, cfg.ApiRateLimit),
		cfg:        cfg,
	}
}

func (c *Client) limitCall(call func()) {
	c.limiter <- struct{}{}
	call()
	<-c.limiter
}
