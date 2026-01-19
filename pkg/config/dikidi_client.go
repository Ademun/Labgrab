package config

import (
	"time"

	"golang.org/x/time/rate"
)

type DikidiClientConfig struct {
	HTTPClientConfig HTTPClientConfig `yaml:"http"`
	SourcesConfig    SourcesConfig    `yaml:"sources"`
}

type HTTPClientConfig struct {
	Timeout        time.Duration `yaml:"timeout"`
	IncreaseFactor float64       `yaml:"increase"`
	DecreaseFactor float64       `yaml:"decrease"`
	MaxRate        rate.Limit    `yaml:"max_rate"`
	MinRate        rate.Limit    `yaml:"min_rate"`
	BurstSize      int           `yaml:"burst"`
}

type SourcesConfig struct {
	SourcesIDsProviderURL string `yaml:"sources_ids_provider"`
	SlotsSourceURL        string `yaml:"slots_source"`
}
