package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.com/berusmenu/api-gateway/internal/http/response"
	authv1 "gitlab.com/berusmenu/proto/gen/go/auth/v1"
	commonv1 "gitlab.com/berusmenu/proto/gen/go/common/v1"
	"google.golang.org/grpc/metadata"
)

const (
	currentUserKey = "current_user"
)

type CurrentUser struct {
	ID            int64
	Role          commonv1.UserRole
	Email         string
	EmailVerified bool
}

func Auth(client authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := ExtractBearerToken(c.GetHeader("Authorization"))
		if !ok {
			response.Error(c, http.StatusUnauthorized, "unauthenticated", "missing or invalid bearer token")
			return
		}

		res, err := client.ValidateToken(c.Request.Context(), &authv1.ValidateTokenRequest{
			AccessToken: token,
		})
		if err != nil {
			response.FromGRPCError(c, err)
			return
		}
		if !res.GetValid() {
			response.Error(c, http.StatusUnauthorized, "unauthenticated", "invalid access token")
			return
		}

		user := CurrentUser{
			ID:            res.GetUserId(),
			Role:          res.GetRole(),
			Email:         res.GetEmail(),
			EmailVerified: res.GetEmailVerified(),
		}

		c.Set(currentUserKey, user)
		c.Request = c.Request.WithContext(WithIdentityMetadata(c.Request.Context(), user))
		c.Next()
	}
}

func ExtractBearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}

	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}

func GetCurrentUser(c *gin.Context) (CurrentUser, bool) {
	value, ok := c.Get(currentUserKey)
	if !ok {
		return CurrentUser{}, false
	}

	user, ok := value.(CurrentUser)
	return user, ok
}

func WithIdentityMetadata(ctx context.Context, user CurrentUser) context.Context {
	return metadata.AppendToOutgoingContext(
		ctx,
		"x-user-id", strconv.FormatInt(user.ID, 10),
		"x-user-role", RoleString(user.Role),
		"x-user-email", user.Email,
		"x-user-email-verified", strconv.FormatBool(user.EmailVerified),
	)
}

func RoleString(role commonv1.UserRole) string {
	switch role {
	case commonv1.UserRole_USER_ROLE_OWNER:
		return "owner"
	case commonv1.UserRole_USER_ROLE_WORKER:
		return "worker"
	case commonv1.UserRole_USER_ROLE_MODERATOR:
		return "moderator"
	default:
		return "unspecified"
	}
}

func RequireVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := GetCurrentUser(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "unauthenticated", "missing authenticated user")
			return
		}

		if !user.EmailVerified {
			response.Error(c, http.StatusForbidden, "email_not_verified", "email verification is required")
			return
		}

		c.Next()
	}
}
