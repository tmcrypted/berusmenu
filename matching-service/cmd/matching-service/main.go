package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/tmcrypted/berusmenu/matching-service/internal/clients"
	"github.com/tmcrypted/berusmenu/matching-service/internal/config"
	"github.com/tmcrypted/berusmenu/matching-service/internal/events"
	"github.com/tmcrypted/berusmenu/matching-service/internal/migrations"
	"github.com/tmcrypted/berusmenu/matching-service/internal/server"
	"github.com/tmcrypted/berusmenu/matching-service/internal/store"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
	inventoryv1 "github.com/tmcrypted/berusmenu/proto/gen/go/inventory/v1"
	matchingv1 "github.com/tmcrypted/berusmenu/proto/gen/go/matching/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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

	inventoryConn, err := grpc.NewClient(cfg.InventoryGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect inventory-service: %v", err)
	}
	defer inventoryConn.Close()
	authConn, err := grpc.NewClient(cfg.AuthGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect auth-service: %v", err)
	}
	defer authConn.Close()

	writer := &kafka.Writer{Addr: kafka.TCP(cfg.KafkaBrokers...), Balancer: &kafka.LeastBytes{}}
	defer writer.Close()

	repo := store.New(pool)
	go events.NewWorker(repo, events.NewKafkaPublisher(writer), cfg).Run(ctx)

	listener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}

	grpcServer := grpc.NewServer()
	matchingv1.RegisterMatchingServiceServer(grpcServer, server.New(server.Dependencies{
		Repo:      repo,
		Inventory: clients.NewInventoryClient(inventoryv1.NewInventoryServiceClient(inventoryConn)),
		Auth:      clients.NewAuthClient(authv1.NewAuthServiceClient(authConn)),
		Events:    events.NewBuilder(cfg),
	}))
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("matching-service listening on %s", cfg.GRPCAddr)
		errCh <- grpcServer.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")
	case err := <-errCh:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("grpc server failed: %v", err)
		}
	}
	grpcServer.GracefulStop()
}
