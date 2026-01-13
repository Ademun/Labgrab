package subscription

import (
	"labgrab/pkg/config"

	"github.com/redis/go-redis/v9"
)

type Deduplicator struct {
	cache *redis.Client
	cfg   *config.DeduplicatorConfig
}

func NewDeduplicator(cache *redis.Client, cfg *config.DeduplicatorConfig) *Deduplicator {
	return &Deduplicator{cache: cache, cfg: cfg}
}
