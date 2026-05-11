package clients

import (
	"context"
	"strings"

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
		UserID:      user.GetId(),
		Email:       user.GetEmail(),
		DisplayName: strings.Join(compact(user.GetLastName(), user.GetFirstName(), user.GetMiddleName()), " "),
		Role:        user.GetRole(),
	}, nil
}

func compact(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
