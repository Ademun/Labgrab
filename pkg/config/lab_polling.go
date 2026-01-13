package config

import "time"

type DeduplicatorConfig struct {
	keyPrefix string        `yaml:"key_prefix"`
	ttl       time.Duration `yaml:"ttl"`
}
