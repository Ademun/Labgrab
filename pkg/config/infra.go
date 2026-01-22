package config

type InfraConfig struct {
	RedisConfig    RedisConfig
	PostgresConfig PostgresConfig
}

type RedisConfig struct {
	Address  string `env:"REDIS_ADDR"`
	Password string `env:"REDIS_PASS"`
	DB       int    `env:"REDIS_DB"`
}

type PostgresConfig struct {
	ConnectionString string `envconfig:"POSTGRES_CONN_STRING"`
}
