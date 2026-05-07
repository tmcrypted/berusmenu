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
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/tmcrypted/berusmenu/auth-service/internal/config"
	"github.com/tmcrypted/berusmenu/auth-service/internal/events"
	"github.com/tmcrypted/berusmenu/auth-service/internal/migrations"
	"github.com/tmcrypted/berusmenu/auth-service/internal/server"
	"github.com/tmcrypted/berusmenu/auth-service/internal/store"
	"github.com/tmcrypted/berusmenu/auth-service/internal/verification"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
	"google.golang.org/grpc"
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

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("ping redis: %v", err)
	}

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBrokers...),
		Topic:    cfg.VerificationTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	listener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}

	authServer := server.New(server.Dependencies{
		Config:       cfg,
		Users:        store.New(pool),
		Verification: verification.New(redisClient, cfg),
		Publisher:    events.NewKafkaPublisher(kafkaWriter, cfg),
	})

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, authServer)
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("auth-service listening on %s", cfg.GRPCAddr)
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
