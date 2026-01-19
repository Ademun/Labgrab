package main

import (
	"context"
	"labgrab/internal/lab_polling"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/subscription"
	"labgrab/internal/user"
	"labgrab/pkg/config"
	"labgrab/pkg/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log := logger.Init()

	log.Info("Starting service")

	log.Info("Loading config")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Fatal error occurred when loading config", "error", err)
	}
	log.Info("Loaded config")

	log.Info("Establishing postgres connection")
	pgconfig, err := pgxpool.ParseConfig(cfg.InfraConfig.PostgresConfig.ConnectionString)
	if err != nil {
		log.Fatal(
			"Fatal error occurred when parsing postgres connection string",
			"conn_string",
			cfg.InfraConfig.PostgresConfig.ConnectionString,
			"error",
			err,
		)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgconfig)
	if err != nil {
		log.Fatal(
			"Fatal error occurred when connecting to postgres server",
			"conn_string",
			cfg.InfraConfig.PostgresConfig.ConnectionString,
			"error",
			err,
		)
	}
	log.Info("Connected to postgresql server")

	log.Info("Establishing redis connection")
	cache := redis.NewClient(&redis.Options{
		Addr:     cfg.InfraConfig.RedisConfig.Address,
		Password: cfg.InfraConfig.RedisConfig.Password,
		DB:       cfg.InfraConfig.RedisConfig.DB,
	})
	log.Info("Connected to redis server")

	log.Info("Setting up dikidi client")
	httpClient := dikidi.NewAdaptiveHTTPClient(&cfg.APIClientConfig.HTTPClientConfig)
	dikidiClient := dikidi.NewClient(&cfg.APIClientConfig, httpClient)
	log.Info("Finished setting up dikidi client")

	log.Info("Setting up polling service")
	slotParser, err := lab_polling.NewParser(&cfg.PollingServiceConfig.ParserConfig)
	if err != nil {
		log.Fatal(
			"Fatal error occurred when creating lab parser",
			"error",
			err,
		)
	}
	pollingService := lab_polling.NewService(dikidiClient, slotParser, log)
	log.Info("Finished setting up polling service")

	log.Info("Setting up subscription service")
	subscriptionRepo := subscription.NewRepo(pool)
	subscriptionService := subscription.NewService(subscriptionRepo, log)
	log.Info("Finished setting up subscription service")

	log.Info("Setting up user service")
	userRepo := user.NewRepo(pool)
	userService := user.NewService(userRepo, log)
	log.Info("Finished setting up user service")

}
