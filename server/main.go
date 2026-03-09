package main

import (
	"context"
	"errors"
	"fmt"
	api_event "labgrab/internal/application/event"
	api_subscription "labgrab/internal/application/subscription"
	api_user "labgrab/internal/application/user"
	"labgrab/internal/application/web"
	"labgrab/internal/auth"
	"labgrab/internal/booking"
	"labgrab/internal/event"
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

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", "error", err)
	}

	tp, err := tracer.Init(ctx, &cfg.InfraConfig.OpenTelemetryConfig)
	if err != nil {
		log.Fatal("Failed to initialize tracer", "error", err)
	}

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal("Failed to shutdown tracer", "error", err)
		}
	}()

	log.Info("Initializing infrastructure")
	pool, cache, err := initInfrastructure(ctx, cfg, log)
	if err != nil {
		log.Fatal("Failed to init infrastructure", "error", err)
	}
	log.Info("Initialized infrastructure")
	defer pool.Close()
	defer cache.Close()

	log.Info("Initializing clients")
	dikidiClient, err := initClients(cfg)
	if err != nil {
		log.Fatal("Failed to initialize clients", "error", err)
	}
	log.Info("Initialized clients")

	log.Info("Initializing services")
	services, err := initServices(ctx, cfg, pool, cache, dikidiClient, log)
	if err != nil {
		log.Fatal("Failed to initialize services", "error", err)
	}
	log.Info("Initialized services")

	log.Info("Initializing HTTP server")
	server, err := initHTTPServer(cfg, pool, services, log)
	if err != nil {
		log.Fatal("Failed to initialize HTTP server", "error", err)
	}
	log.Info("Initialized HTTP server")

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

func initClients(cfg *config.Config) (*dikidi.Client, error) {
	parser, err := dikidi.NewParser(&cfg.PollingServiceConfig.ParserConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup client parser: %w", err)
	}
	dikidiClient := dikidi.NewClient(&cfg.APIClientConfig, parser)
	return dikidiClient, nil
}

type Services struct {
	Subscription *subscription.Service
	User         *user.Service
	Auth         *auth.Service
	Event        *event.Service
	Booking      *booking.Service
	Telegram     *telegram.Service
}

func initServices(
	ctx context.Context,
	cfg *config.Config,
	pool *pgxpool.Pool,
	cache *redis.Client,
	dikidiClient *dikidi.Client,
	logger *zap.SugaredLogger,
) (*Services, error) {
	eventService := event.NewService(dikidiClient)

	subscriptionRepo := subscription.NewRepo(pool)
	deduplicator := subscription.NewDeduplicator(cache, cfg.SubscriptionServiceConfig.DeduplicatorConfig)
	subscriptionService := subscription.NewService(subscriptionRepo, deduplicator)

	userRepo := user.NewRepo(pool)
	userService := user.NewService(userRepo)

	authRepo := auth.NewRepo(pool)
	authService, err := auth.NewService(authRepo, cache, dikidiClient, &cfg.AuthServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup auth service: %w", err)
	}

	telegramService, err := telegram.NewService(&cfg.TelegramConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to set up telegram service: %w", err)
	}
	go telegramService.Start(ctx)

	bookingRepo := booking.NewRepo(pool)
	bookingService := booking.NewService(bookingRepo, dikidiClient)

	eventScheduler := api_event.NewScheduler(
		dikidiClient,
		eventService,
		subscriptionService,
		userService,
		authService,
		bookingService,
		telegramService,
		logger,
	)
	if err := eventScheduler.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start event scheduler: %w", err)
	}

	return &Services{
		Subscription: subscriptionService,
		User:         userService,
		Auth:         authService,
		Event:        eventService,
		Booking:      bookingService,
		Telegram:     telegramService,
	}, nil
}

func initHTTPServer(cfg *config.Config, pool *pgxpool.Pool, services *Services, log *zap.SugaredLogger) (*http.Server, error) {
	r := mux.NewRouter()

	log.Info("Setting up user domain routes")
	userHandler := api_user.NewHandler(services.Auth, services.User, log)
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
