// Command server starts the Connect-protocol HTTP server for the Habit Tracker.
// The Connect protocol (https://connectrpc.com) speaks HTTP/1.1 + HTTP/2 with
// JSON or binary protobuf bodies — browsers can call it directly without Envoy.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	pb "github.com/yourorg/habit-tracker/gen/habit/v1"
	"github.com/yourorg/habit-tracker/gen/habit/v1/habitv1connect"
	"github.com/yourorg/habit-tracker/internal/auth"
	"github.com/yourorg/habit-tracker/internal/config"
	"github.com/yourorg/habit-tracker/internal/handler"
	"github.com/yourorg/habit-tracker/internal/middleware"
	repopostgres "github.com/yourorg/habit-tracker/internal/repository/postgres"
	"github.com/yourorg/habit-tracker/internal/service"
	pkglogger "github.com/yourorg/habit-tracker/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	logger := pkglogger.New(cfg.Server.Env)
	log.Logger = logger
	logger.Info().Str("env", cfg.Server.Env).Int("port", cfg.Server.Port).Msg("starting habit-tracker")

	// ── Database ────────────────────────────────────────────────────────────

	gormCfg := &gorm.Config{}
	if cfg.Server.Env == "development" {
		gormCfg.Logger = gormlogger.Default.LogMode(gormlogger.Info)
	} else {
		gormCfg.Logger = gormlogger.Default.LogMode(gormlogger.Error)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), gormCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.Database.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxConns / 2)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	defer sqlDB.Close()

	// ── Migrations ──────────────────────────────────────────────────────────

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.Name, cfg.Database.SSLMode,
	)
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init migrations")
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatal().Err(err).Msg("migration failed")
	}
	logger.Info().Msg("migrations applied")

	// ── Repositories ────────────────────────────────────────────────────────

	userRepo         := repopostgres.NewUserRepository(db)
	habitRepo        := repopostgres.NewHabitRepository(db)
	habitLogRepo     := repopostgres.NewHabitLogRepository(db)
	refreshTokenRepo := repopostgres.NewRefreshTokenRepository(db)
	settingsRepo     := repopostgres.NewSettingsRepository(db)

	// ── Services ────────────────────────────────────────────────────────────

	tokenSvc   := auth.NewTokenService(cfg.JWT.Secret)
	authSvc    := service.NewAuthService(userRepo, refreshTokenRepo, settingsRepo, tokenSvc)
	habitSvc   := service.NewHabitService(habitRepo, habitLogRepo)
	progressSvc:= service.NewProgressService(habitRepo, habitLogRepo)
	statsSvc   := service.NewStatsService(habitRepo, habitLogRepo)
	settingsSvc:= service.NewSettingsService(settingsRepo)

	// ── Connect interceptors ────────────────────────────────────────────────

	// Auth interceptor wraps each RPC; public methods are listed inside.
	authInterceptor := middleware.NewConnectAuthInterceptor(tokenSvc)
	logInterceptor  := middleware.NewConnectLoggingInterceptor(logger)

	interceptors := connect.WithInterceptors(logInterceptor, authInterceptor)

	// ── Mount Connect handlers onto a plain ServeMux ────────────────────────
	// Each service gets a path prefix: /habit.v1.AuthService/MethodName
	// This is the Connect protocol URL scheme — no Envoy required.

	mux := http.NewServeMux()

	mux.Handle(habitv1connect.NewAuthServiceHandler(
		handler.NewAuthHandler(authSvc),
		interceptors,
	))
	mux.Handle(habitv1connect.NewHabitServiceHandler(
		handler.NewHabitHandler(habitSvc),
		interceptors,
	))
	mux.Handle(habitv1connect.NewProgressServiceHandler(
		handler.NewProgressHandler(progressSvc),
		interceptors,
	))
	mux.Handle(habitv1connect.NewStatsServiceHandler(
		handler.NewStatsHandler(statsSvc),
		interceptors,
	))
	mux.Handle(habitv1connect.NewSettingsServiceHandler(
		handler.NewSettingsHandler(settingsSvc),
		interceptors,
	))

	// Health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// ── CORS ─────────────────────────────────────────────────────────────────
	// connectcors.AllowedHeaders() returns the headers Connect needs beyond defaults.

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins(cfg.Server.Env),
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
		MaxAge:         7200,
	})

	// ── HTTP server ──────────────────────────────────────────────────────────
	// h2c lets HTTP/2 work without TLS (required for Connect binary protocol).
	// JSON Connect works over plain HTTP/1.1 too — both are supported.

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      corsHandler.Handler(h2c.NewHandler(mux, &http2.Server{})),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Graceful shutdown ────────────────────────────────────────────────────

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info().Str("addr", httpServer.Addr).Msg("Connect server listening")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	logger.Info().Msg("shutting down")

	timeout := time.Duration(cfg.Server.GracefulTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("forced shutdown")
		_ = httpServer.Close()
	}
	logger.Info().Msg("server stopped")
}

func allowedOrigins(env string) []string {
	if env == "development" {
		return []string{"http://localhost:5173", "http://localhost:4173", "http://localhost:3000"}
	}
	// In production, set CORS_ALLOWED_ORIGINS env var or hardcode your domain.
	origin := os.Getenv("CORS_ALLOWED_ORIGINS")
	if origin == "" {
		return []string{"https://yourdomain.com"}
	}
	return []string{origin}
}

// Ensure pb is used (handler package uses it; this import is for the gen path check).
var _ = pb.AuthService_ServiceDesc
