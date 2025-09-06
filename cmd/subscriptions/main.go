package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/oziev02/subscriptions-service/internal/adapters/httpapi"
	"github.com/oziev02/subscriptions-service/internal/adapters/repo/postgres"
	"github.com/oziev02/subscriptions-service/internal/pkg/config"
	"github.com/oziev02/subscriptions-service/internal/pkg/logger"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	log := logger.New(cfg.LogLevel)
	defer log.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.DB.DSN)
	if err != nil {
		log.Fatal("db connect", zap.Error(err))
	}
	defer pool.Close()

	repo := postgres.NewSubscriptionRepo(pool, log)
	api := httpapi.NewServer(cfg, log, repo)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           api.Router(),
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("http listening", zap.Int("port", cfg.HTTP.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("shutting down...")
	ctxShut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctxShut)
	log.Info("bye")
	_ = os.Stderr.Sync()
}
