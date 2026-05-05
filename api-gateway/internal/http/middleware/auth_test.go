package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	authv1 "gitlab.com/berusmenu/proto/gen/go/auth/v1"
	commonv1 "gitlab.com/berusmenu/proto/gen/go/common/v1"
	"google.golang.org/grpc"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantOK    bool
	}{
		{name: "valid", header: "Bearer abc", wantToken: "abc", wantOK: true},
		{name: "case insensitive", header: "bearer abc", wantToken: "abc", wantOK: true},
		{name: "trims token", header: "Bearer   abc  ", wantToken: "abc", wantOK: true},
		{name: "missing", header: "", wantOK: false},
		{name: "wrong scheme", header: "Basic abc", wantOK: false},
		{name: "empty token", header: "Bearer ", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, gotOK := ExtractBearerToken(tt.header)
			if gotOK != tt.wantOK {
				t.Fatalf("ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotToken != tt.wantToken {
				t.Fatalf("token = %q, want %q", gotToken, tt.wantToken)
			}
		})
	}
}

func TestAuthMiddlewareStoresCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
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
	router.Use(RequestID(), Auth(client))
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
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if client.validateToken != "token" {
		t.Fatalf("validated token = %q, want token", client.validateToken)
	}
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID(), Auth(&authClientStub{}))
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
