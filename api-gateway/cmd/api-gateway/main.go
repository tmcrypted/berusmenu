package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/config"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/grpcclients"
	httpserver "github.com/tmcrypted/berusmenu/api-gateway/internal/http"
)

func main() {
	cfg := config.Load()

	if cfg.AppEnv == "release" || cfg.AppEnv == "prod" || cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	clients, err := grpcclients.New(cfg)
	if err != nil {
		log.Fatalf("init grpc clients: %v", err)
	}
	defer clients.Close()

	router := httpserver.NewRouter(httpserver.Dependencies{
		Config:        cfg,
		Auth:          clients.Auth,
		Inventory:     clients.Inventory,
		Matching:      clients.Matching,
		Communication: clients.Communication,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("api-gateway listening on %s", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stopCh:
		log.Printf("shutdown signal received: %s", sig)
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server failed: %v", err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
}
