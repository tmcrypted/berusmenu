package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tmcrypted/berusmenu/notification-service/internal/config"
	"github.com/tmcrypted/berusmenu/notification-service/internal/consumer"
	notificationservice "github.com/tmcrypted/berusmenu/notification-service/internal/service"
)

func main() {
	cfg := config.Load()

	service := notificationservice.New(notificationservice.NewLogSender())
	c := consumer.New(cfg, service)
	defer c.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
