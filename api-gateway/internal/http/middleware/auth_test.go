package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/config"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	"google.golang.org/grpc"
)

func TestAuthMiddlewareStoresCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testConfig()
	client := &authClientStub{
		validateResp: &authv1.ValidateTokenResponse{
			Valid:         true,
			UserId:        42,
			Role:          commonv1.UserRole_USER_ROLE_OWNER,
			Email:         "owner@example.com",
			EmailVerified: true,
		},
	}

	router := gin.New()
	router.Use(RequestID(), Auth(client, cfg))
	router.GET("/private", func(c *gin.Context) {
		user, ok := GetCurrentUser(c)
		if !ok {
			t.Fatal("current user was not stored")
		}
		if user.ID != 42 || user.Role != commonv1.UserRole_USER_ROLE_OWNER || user.Email != "owner@example.com" || !user.EmailVerified {
			t.Fatalf("unexpected user: %+v", user)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.AddCookie(&http.Cookie{Name: cfg.AccessCookieName, Value: "token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if client.validateToken != "token" {
		t.Fatalf("validated token = %q, want token", client.validateToken)
	}
}

func TestAuthMiddlewareRejectsBearerWithoutCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID(), Auth(&authClientStub{}, testConfig()))
	router.GET("/private", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID(), Auth(&authClientStub{}, testConfig()))
	router.GET("/private", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func testConfig() config.Config {
	return config.Config{
		AccessCookieName:  "berusmenu_access",
		RefreshCookieName: "berusmenu_refresh",
		CSRFCookieName:    "berusmenu_csrf",
		CSRFHeaderName:    "X-CSRF-Token",
	}
}

func TestRequireVerifiedRejectsUnverifiedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(currentUserKey, CurrentUser{ID: 1, EmailVerified: false})
		c.Next()
	}, RequireVerified())
	router.GET("/verified", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/verified", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

type authClientStub struct {
	authv1.AuthServiceClient
	validateResp  *authv1.ValidateTokenResponse
	validateErr   error
	validateToken string
}

func (s *authClientStub) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest, opts ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
	s.validateToken = req.GetAccessToken()
	if s.validateErr != nil {
		return nil, s.validateErr
	}
	return s.validateResp, nil
}
