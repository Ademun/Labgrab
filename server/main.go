package main

import (
	"context"
	"errors"
	"fmt"
	api_subscription "labgrab/internal/application/subscription"
	api_user "labgrab/internal/application/user"
	"labgrab/internal/application/web"
	"labgrab/internal/auth"
	"labgrab/internal/lab_polling"
	"labgrab/internal/shared/api/dikidi"
	"labgrab/internal/shared/routing"
	"labgrab/internal/subscription"
	"labgrab/internal/telegram"
	"labgrab/internal/user"
	"labgrab/pkg/config"
	"labgrab/pkg/logger"
	"labgrab/pkg/tracer"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

	tp, err := tracer.Init(ctx, &cfg.InfraConfig.OpenTelemetryConfig)
	if err != nil {
		log.Fatal("Fatal error occurred when initializing tracer", "error", err)
	}

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal("Fatal error occurred when shutting down tracer", "error", err)
		}
	}()

	pool, cache, err := initInfrastructure(ctx, cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize infrastructure", "error", err)
	}
	defer pool.Close()
	defer cache.Close()

	dikidiClient, err := initClients(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize clients", "error", err)
	}

	services, err := initServices(ctx, cfg, pool, cache, dikidiClient, log)
	if err != nil {
		log.Fatal("Failed to initialize services", "error", err)
	}

	go func() {
		log.Info("Starting Telegram bot")
		services.Telegram.Start(ctx)
	}()

	server, err := initHTTPServer(cfg, pool, services, log)
	if err != nil {
		log.Fatal("Failed to initialize HTTP server", "error", err)
	}

	go func() {
		log.Info("Starting HTTP server on 127.0.0.1:8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("HTTP server error", "error", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Server shutdown error", "error", err)
	}

	log.Info("Server stopped")
}

func initInfrastructure(ctx context.Context, cfg *config.Config, log *zap.SugaredLogger) (*pgxpool.Pool, *redis.Client, error) {
	log.Info("Establishing postgres connection")
	pgconfig, err := pgxpool.ParseConfig(cfg.InfraConfig.PostgresConfig.ConnectionString)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing postgres connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	log.Info("Connected to postgresql server")

	log.Info("Establishing redis connection")
	cache := redis.NewClient(&redis.Options{
		Addr:     cfg.InfraConfig.RedisConfig.Address,
		Password: cfg.InfraConfig.RedisConfig.Password,
		DB:       cfg.InfraConfig.RedisConfig.DB,
	})
	if err := cache.Ping(ctx).Err(); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("connecting to redis: %w", err)
	}
	log.Info("Connected to redis server")

	return pool, cache, nil
}

func initClients(cfg *config.Config, log *zap.SugaredLogger) (*dikidi.Client, error) {
	log.Info("Setting up dikidi client")
	httpClient := dikidi.NewAdaptiveHTTPClient(&cfg.APIClientConfig.HTTPClientConfig)
	dikidiClient := dikidi.NewClient(&cfg.APIClientConfig, httpClient)
	log.Info("Finished setting up dikidi client")
	return dikidiClient, nil
}

type Services struct {
	Subscription *subscription.Service
	User         *user.Service
	Auth         *auth.Service
	LabPolling   *lab_polling.Service
	Telegram     *telegram.Service
}

func initServices(
	ctx context.Context,
	cfg *config.Config,
	pool *pgxpool.Pool,
	cache *redis.Client,
	dikidiClient *dikidi.Client,
	log *zap.SugaredLogger,
) (*Services, error) {
	log.Info("Setting up polling service")
	slotParser, err := lab_polling.NewParser(&cfg.PollingServiceConfig.ParserConfig)
	if err != nil {
		return nil, fmt.Errorf("creating lab parser: %w", err)
	}
	labPollingService := lab_polling.NewService(dikidiClient, slotParser, log)
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
	authService := auth.NewService(cache, &cfg.AuthServiceConfig, log)
	log.Info("Finished setting up auth service")

	log.Info("Setting up telegram service")
	telegramService, err := telegram.NewService(&cfg.TelegramConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to set up telegram service: %w", err)
	}
	log.Info("Finished setting up telegram service")

	log.Info("Setting up schedulers")
	subscriptionScheduler := api_subscription.NewScheduler(dikidiClient, labPollingService, subscriptionService, userService, telegramService, log)
	if err := subscriptionScheduler.Start(ctx); err != nil {
		return nil, fmt.Errorf("starting subscription scheduler: %w", err)
	}

	return &Services{
		Subscription: subscriptionService,
		User:         userService,
		Auth:         authService,
		LabPolling:   labPollingService,
		Telegram:     telegramService,
	}, nil
}

func initHTTPServer(cfg *config.Config, pool *pgxpool.Pool, services *Services, log *zap.SugaredLogger) (*http.Server, error) {
	log.Info("Setting up routes")
	r := mux.NewRouter()

	log.Info("Setting up user domain routes")
	userHandler := api_user.NewHandler(services.Auth, services.User, pool, log)
	userHandler.RegisterRoutes(r)
	log.Info("Finished setting up user domain routes")

	log.Info("Setting up subscription domain routes")
	subscriptionHandler := api_subscription.NewHandler(services.Auth, services.Subscription, log)
	subscriptionHandler.RegisterRoutes(r)
	log.Info("Finished setting up subscription domain routes")

	log.Info("Setting up web domain routes")
	webHandler := web.NewHandler(log)
	webHandler.RegisterRoutes(r)
	log.Info("Finished setting up web domain routes")

	server := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: routing.CORSMiddleware(r),
	}

	return server, nil
}
