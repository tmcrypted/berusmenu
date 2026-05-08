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
	"github.com/tmcrypted/berusmenu/inventory-service/internal/config"
	inventorymatching "github.com/tmcrypted/berusmenu/inventory-service/internal/matching"
	"github.com/tmcrypted/berusmenu/inventory-service/internal/migrations"
	"github.com/tmcrypted/berusmenu/inventory-service/internal/server"
	"github.com/tmcrypted/berusmenu/inventory-service/internal/store"
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

	matchingConn, err := grpc.NewClient(cfg.MatchingGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect matching-service: %v", err)
	}
	defer matchingConn.Close()

	listener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}

	inventoryServer := server.New(server.Dependencies{
		Pvz:             store.New(pool),
		ActiveVacancies: inventorymatching.NewActiveVacancyClient(matchingv1.NewMatchingServiceClient(matchingConn)),
	})

	grpcServer := grpc.NewServer()
	inventoryv1.RegisterInventoryServiceServer(grpcServer, inventoryServer)
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("inventory-service listening on %s", cfg.GRPCAddr)
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
