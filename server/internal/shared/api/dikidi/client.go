package dikidi

import (
	"labgrab/pkg/config"
)

type Client struct {
	parser     *Parser
	serviceIDs []int
	cfg        *config.DikidiClientConfig
}

func NewClient(cfg *config.DikidiClientConfig, parser *Parser) *Client {
	return &Client{
		serviceIDs: make([]int, 0),
		parser:     parser,
		cfg:        cfg,
	}
}
