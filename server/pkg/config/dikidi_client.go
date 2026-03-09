package config

type DikidiClientConfig struct {
	ServiceProviderURL string `yaml:"service_provider_url"`
	EventProviderURL   string `yaml:"event_provider_url"`
	ApiRateLimit       int    `yaml:"api_rate_limit"`
}
