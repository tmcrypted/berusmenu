package matching

import (
	"context"

	matchingv1 "github.com/tmcrypted/berusmenu/proto/gen/go/matching/v1"
	"google.golang.org/grpc/metadata"
)

type ActiveVacancyClient struct {
	client matchingv1.MatchingServiceClient
}

func NewActiveVacancyClient(client matchingv1.MatchingServiceClient) *ActiveVacancyClient {
	return &ActiveVacancyClient{client: client}
}

func (c *ActiveVacancyClient) HasActiveVacancyForPvz(ctx context.Context, pvzID int64) (bool, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ctx = metadata.NewOutgoingContext(ctx, md.Copy())
	}
	res, err := c.client.HasActiveVacancyForPvz(ctx, &matchingv1.HasActiveVacancyForPvzRequest{PvzId: pvzID})
	if err != nil {
		return false, err
	}
	return res.GetHasActiveVacancy(), nil
}
