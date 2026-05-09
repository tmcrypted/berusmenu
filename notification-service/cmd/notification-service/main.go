package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmcrypted/berusmenu/notification-service/internal/config"
	"github.com/tmcrypted/berusmenu/notification-service/internal/consumer"
	"github.com/tmcrypted/berusmenu/notification-service/internal/migrations"
	notificationservice "github.com/tmcrypted/berusmenu/notification-service/internal/service"
	"github.com/tmcrypted/berusmenu/notification-service/internal/store"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping postgres: %v", err)
	}
	if err := migrations.Apply(ctx, pool, cfg.MigrationsDir); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	service := notificationservice.New(notificationservice.NewLogSender())
	c := consumer.NewWithDeduplicator(cfg, service, nil, store.New(pool))
	defer c.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- c.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			log.Fatalf("consumer failed: %v", err)
		}
	}

	// Give it a moment to commit offsets
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("shutting down gracefully...")
}
