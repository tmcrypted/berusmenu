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
	"gitlab.com/berusmenu/auth-service/internal/config"
	"gitlab.com/berusmenu/auth-service/internal/events"
	"gitlab.com/berusmenu/auth-service/internal/migrations"
	"gitlab.com/berusmenu/auth-service/internal/server"
	"gitlab.com/berusmenu/auth-service/internal/store"
	"gitlab.com/berusmenu/auth-service/internal/verification"
	authv1 "gitlab.com/berusmenu/proto/gen/go/auth/v1"
	"google.golang.org/grpc"
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
