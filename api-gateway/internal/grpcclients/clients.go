package grpcclients

import (
	"fmt"

	"github.com/tmcrypted/berusmenu/api-gateway/internal/config"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
	communicationv1 "github.com/tmcrypted/berusmenu/proto/gen/go/communication/v1"
	inventoryv1 "github.com/tmcrypted/berusmenu/proto/gen/go/inventory/v1"
	matchingv1 "github.com/tmcrypted/berusmenu/proto/gen/go/matching/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Auth          authv1.AuthServiceClient
	Inventory     inventoryv1.InventoryServiceClient
	Matching      matchingv1.MatchingServiceClient
	Communication communicationv1.CommunicationServiceClient

	conns []*grpc.ClientConn
}

func New(cfg config.Config) (*Clients, error) {
	authConn, err := newConn(cfg.AuthGRPCAddr)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	inventoryConn, err := newConn(cfg.InventoryGRPCAddr)
	if err != nil {
		authConn.Close()
		return nil, fmt.Errorf("inventory: %w", err)
	}

	matchingConn, err := newConn(cfg.MatchingGRPCAddr)
	if err != nil {
		authConn.Close()
		inventoryConn.Close()
		return nil, fmt.Errorf("matching: %w", err)
	}

	communicationConn, err := newConn(cfg.CommunicationGRPCAddr)
	if err != nil {
		authConn.Close()
		inventoryConn.Close()
		matchingConn.Close()
		return nil, fmt.Errorf("communication: %w", err)
	}

	return &Clients{
		Auth:          authv1.NewAuthServiceClient(authConn),
		Inventory:     inventoryv1.NewInventoryServiceClient(inventoryConn),
		Matching:      matchingv1.NewMatchingServiceClient(matchingConn),
		Communication: communicationv1.NewCommunicationServiceClient(communicationConn),
		conns: []*grpc.ClientConn{
			authConn,
			inventoryConn,
			matchingConn,
			communicationConn,
		},
	}, nil
}

func (c *Clients) Close() {
	if c == nil {
		return
	}

	for _, conn := range c.conns {
		if conn != nil {
			conn.Close()
		}
	}
}

func newConn(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
