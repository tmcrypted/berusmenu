package clients

import (
	"context"

	"github.com/tmcrypted/berusmenu/matching-service/internal/dto"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
)

type AuthClient struct {
	client authv1.AuthServiceClient
}

func NewAuthClient(client authv1.AuthServiceClient) *AuthClient {
	return &AuthClient{client: client}
}

func (c *AuthClient) GetUserByID(ctx context.Context, userID int64) (dto.UserSnapshot, error) {
	res, err := c.client.GetUserByID(withForwardedIdentity(ctx), &authv1.GetUserByIDRequest{UserId: userID})
	if err != nil {
		return dto.UserSnapshot{}, err
	}
	user := res.GetUser()
	return dto.UserSnapshot{
		UserID:   user.GetId(),
		Email:    user.GetEmail(),
		FullName: user.GetFullName(),
		Role:     user.GetRole(),
	}, nil
}
