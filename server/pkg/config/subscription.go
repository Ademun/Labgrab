package config

import "time"

type SubscriptionServiceConfig struct {
	DeduplicatorConfig *DeduplicatorConfig `yaml:"deduplicator"`
}

type DeduplicatorConfig struct {
	KeyPrefix string        `yaml:"key_prefix"`
	TTL       time.Duration `yaml:"ttl"`
}
