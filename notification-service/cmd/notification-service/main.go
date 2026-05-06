package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/berusmenu/notification-service/internal/config"
	"gitlab.com/berusmenu/notification-service/internal/consumer"
)

func main() {
	cfg := config.Load()
	
	c := consumer.New(cfg)
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
