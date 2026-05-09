package clients

import (
	"context"

	"github.com/tmcrypted/berusmenu/matching-service/internal/dto"
	inventoryv1 "github.com/tmcrypted/berusmenu/proto/gen/go/inventory/v1"
)

type InventoryClient struct {
	client inventoryv1.InventoryServiceClient
}

func NewInventoryClient(client inventoryv1.InventoryServiceClient) *InventoryClient {
	return &InventoryClient{client: client}
}

func (c *InventoryClient) CheckPvzForVacancy(ctx context.Context, pvzID int64) (dto.PvzSnapshot, error) {
	res, err := c.client.CheckPvzForVacancy(withForwardedIdentity(ctx), &inventoryv1.CheckPvzForVacancyRequest{PvzId: pvzID})
	if err != nil {
		return dto.PvzSnapshot{}, err
	}
	pvz := res.GetPvz()
	return dto.PvzSnapshot{
		PvzID:       pvz.GetPvzId(),
		OwnerID:     pvz.GetOwnerId(),
		CompanyName: pvz.GetCompanyName(),
		Address:     pvz.GetAddress(),
	}, nil
}
