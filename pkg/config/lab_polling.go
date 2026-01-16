package config

import "time"

type DeduplicatorConfig struct {
	KeyPrefix string        `yaml:"key_prefix"`
	TTL       time.Duration `yaml:"ttl"`
}
