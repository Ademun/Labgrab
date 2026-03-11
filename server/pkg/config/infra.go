package config

type InfraConfig struct {
	RedisConfig    RedisConfig
	PostgresConfig PostgresConfig
}

type RedisConfig struct {
	Address  string `envconfig:"REDIS_ADDR"`
	Password string `envconfig:"REDIS_PASS"`
	DB       int    `envconfig:"REDIS_DB"`
}

type PostgresConfig struct {
	ConnectionString string `envconfig:"POSTGRES_CONN_STRING"`
}
