package config

type Config struct {
	APIClientConfig           APIClientConfig           `yaml:"api_client"`
	SubscriptionServiceConfig SubscriptionServiceConfig `yaml:"subscription_service"`
}
