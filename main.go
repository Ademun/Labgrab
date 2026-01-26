package main

import (
	"context"
	api_subscription "labgrab/internal/application/subscription"
	api_user "labgrab/internal/application/user"
	user_usecase "labgrab/internal/application/user/usecase"
	"labgrab/internal/auth"
	"labgrab/internal/lab_polling"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/subscription"
	"labgrab/internal/user"
	"labgrab/pkg/config"
	"labgrab/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
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
	_ = lab_polling.NewService(dikidiClient, slotParser, log)
	log.Info("Finished setting up polling service")

	log.Info("Setting up subscription service")
	subscriptionRepo := subscription.NewRepo(pool)
	deduplicator := subscription.NewDeduplicator(cache, cfg.SubscriptionServiceConfig.DeduplicatorConfig)
	subscriptionService := subscription.NewService(subscriptionRepo, deduplicator, log)
	log.Info("Finished setting up subscription service")

	log.Info("Setting up user service")
	userRepo := user.NewRepo(pool)
	userService := user.NewService(userRepo, log)
	log.Info("Finished setting up user service")

	log.Info("Setting up auth service")
	authService := auth.NewService(&cfg.AuthServiceConfig, log)
	log.Info("Finished setting up auth service")

	log.Info("Setting up routes")
	r := mux.NewRouter()
	log.Info("Setting up user domain routes")
	authUserUseCase := user_usecase.NewAuthUserUseCase(authService, userService)
	newUserUseCase := user_usecase.NewNewUserUseCase(userService, subscriptionService)
	userHandler := api_user.NewHandler(authUserUseCase, newUserUseCase)
	r.HandleFunc("/api/user/auth", userHandler.Auth).Methods(http.MethodPost)
	r.HandleFunc("/api/user/new", userHandler.NewUser).Methods(http.MethodPost)
	log.Info("Finished setting up user domain routes")
	log.Info("Setting up subscription domain routes")
	subscriptionHandler := api_subscription.NewHandler(subscriptionService, log)
	subscriptionHandler.RegisterRoutes(r)
	log.Info("Finished setting up subscription domain routes")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Failed to start http server", "error", err)
	}
	log.Info("Finished setting up routes")
	select {
	case <-ctx.Done():
		log.Info("Shutting down server")
		return
	}
}
