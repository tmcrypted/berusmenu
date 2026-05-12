package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/config"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/middleware"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/mock"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestParseUserRole(t *testing.T) {
	tests := []struct {
		value string
		want  commonv1.UserRole
		ok    bool
	}{
		{value: "", want: commonv1.UserRole_USER_ROLE_WORKER, ok: true},
		{value: "worker", want: commonv1.UserRole_USER_ROLE_WORKER, ok: true},
		{value: "owner", want: commonv1.UserRole_USER_ROLE_OWNER, ok: true},
		{value: "moderator", want: commonv1.UserRole_USER_ROLE_MODERATOR, ok: true},
		{value: "bad", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := ParseUserRole(tt.value)
			if tt.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatal("expected error")
			}
			if tt.ok && got != tt.want {
				t.Fatalf("role = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRegisterSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().Register(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *authv1.RegisterRequest, opts ...grpc.CallOption) (*authv1.RegisterResponse, error) {
			if req.GetRole() != commonv1.UserRole_USER_ROLE_OWNER {
				t.Errorf("role = %s, want owner", req.GetRole())
			}
			return &authv1.RegisterResponse{Auth: sampleAuthPayload(false)}, nil
		},
	)

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPost, "/v1/auth/register", `{
		"email":"owner@example.com",
		"password":"password123",
		"role":"owner"
	}`, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "access_token") || strings.Contains(rec.Body.String(), "refresh_token") {
		t.Fatalf("response leaked tokens: %s", rec.Body.String())
	}
	if cookieValue(rec, "berusmenu_access") != "access" || cookieValue(rec, "berusmenu_refresh") != "refresh" {
		t.Fatalf("auth cookies were not set: %v", rec.Result().Cookies())
	}
}

func TestRegisterValidationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPost, "/v1/auth/register", `{
		"email":"bad",
		"password":"short"
	}`, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLoginWrongCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, status.Error(codes.Unauthenticated, "bad credentials"))

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPost, "/v1/auth/login", `{
		"email":"worker@example.com",
		"password":"wrong"
	}`, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"unauthenticated"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestResendVerificationCodeRequiresToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPost, "/v1/auth/resend-code", `{}`, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestResendVerificationCodeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().ValidateToken(gomock.Any(), gomock.Any()).Return(&authv1.ValidateTokenResponse{
		Valid:         true,
		UserId:        7,
		Role:          commonv1.UserRole_USER_ROLE_WORKER,
		Email:         "worker@example.com",
		EmailVerified: false,
	}, nil)

	client.EXPECT().ResendVerificationCode(gomock.Any(), gomock.Any()).Return(&authv1.ResendVerificationCodeResponse{
		CodeExpiresAt:     timestamppb.Now(),
		ResendAvailableAt: timestamppb.Now(),
	}, nil)

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPost, "/v1/auth/resend-code", `{}`, "access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestVerifyEmailSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().ValidateToken(gomock.Any(), gomock.Any()).Return(&authv1.ValidateTokenResponse{
		Valid:         true,
		UserId:        7,
		Role:          commonv1.UserRole_USER_ROLE_WORKER,
		Email:         "worker@example.com",
		EmailVerified: false,
	}, nil)

	client.EXPECT().VerifyEmail(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *authv1.VerifyEmailRequest, opts ...grpc.CallOption) (*authv1.VerifyEmailResponse, error) {
			if req.GetCode() != "123456" {
				t.Errorf("code = %q, want 123456", req.GetCode())
			}
			return &authv1.VerifyEmailResponse{User: sampleUser(true)}, nil
		},
	)

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPost, "/v1/auth/verify-email", `{"code":"123456"}`, "access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestRefreshReadsCookieAndRotatesCookies(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().Refresh(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *authv1.RefreshRequest, opts ...grpc.CallOption) (*authv1.RefreshResponse, error) {
			if req.GetRefreshToken() != "old-refresh" {
				t.Errorf("refresh_token = %q, want old-refresh", req.GetRefreshToken())
			}
			return &authv1.RefreshResponse{Auth: sampleAuthPayload(true)}, nil
		},
	)

	router := newTestRouter(client)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "berusmenu_refresh", Value: "old-refresh"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if cookieValue(rec, "berusmenu_access") != "access" || cookieValue(rec, "berusmenu_refresh") != "refresh" {
		t.Fatalf("rotated cookies were not set: %v", rec.Result().Cookies())
	}
}

func TestLogoutClearsCookies(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().Logout(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *authv1.LogoutRequest, opts ...grpc.CallOption) (*authv1.LogoutResponse, error) {
			if req.GetRefreshToken() != "refresh" {
				t.Errorf("refresh_token = %q, want refresh", req.GetRefreshToken())
			}
			return &authv1.LogoutResponse{}, nil
		},
	)

	router := newTestRouter(client)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "berusmenu_refresh", Value: "refresh"})
	req.AddCookie(&http.Cookie{Name: "berusmenu_csrf", Value: "csrf"})
	req.Header.Set("X-CSRF-Token", "csrf")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if maxAgeForCookie(rec, "berusmenu_access") >= 0 || maxAgeForCookie(rec, "berusmenu_refresh") >= 0 {
		t.Fatalf("auth cookies were not cleared: %v", rec.Result().Cookies())
	}
}

func TestProtectedUnsafeEndpointRequiresCSRF(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().ValidateToken(gomock.Any(), gomock.Any()).Return(&authv1.ValidateTokenResponse{
		Valid:         true,
		UserId:        7,
		Role:          commonv1.UserRole_USER_ROLE_WORKER,
		Email:         "worker@example.com",
		EmailVerified: true,
	}, nil)

	router := newTestRouter(client)
	req := httptest.NewRequest(http.MethodPatch, "/v1/me", strings.NewReader(`{"first_name":"Ivan"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "berusmenu_access", Value: "access"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestGetMeRequiresToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodGet, "/v1/me", "", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetMeRejectsUnverifiedUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().ValidateToken(gomock.Any(), gomock.Any()).Return(&authv1.ValidateTokenResponse{
		Valid:         true,
		UserId:        7,
		Role:          commonv1.UserRole_USER_ROLE_WORKER,
		Email:         "worker@example.com",
		EmailVerified: false,
	}, nil)

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodGet, "/v1/me", "", "access")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestGetSessionAllowsUnverifiedUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().ValidateToken(gomock.Any(), gomock.Any()).Return(&authv1.ValidateTokenResponse{
		Valid:         true,
		UserId:        7,
		Role:          commonv1.UserRole_USER_ROLE_WORKER,
		Email:         "worker@example.com",
		EmailVerified: false,
	}, nil)

	client.EXPECT().GetUser(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *authv1.GetUserRequest, opts ...grpc.CallOption) (*authv1.GetUserResponse, error) {
			if req.GetUserId() != 7 {
				t.Errorf("user_id = %d, want 7", req.GetUserId())
			}
			return &authv1.GetUserResponse{User: sampleUser(false)}, nil
		},
	)

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodGet, "/v1/auth/session", "", "access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestGetMeSuccessUsesIdentityMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().ValidateToken(gomock.Any(), gomock.Any()).Return(&authv1.ValidateTokenResponse{
		Valid:         true,
		UserId:        7,
		Role:          commonv1.UserRole_USER_ROLE_WORKER,
		Email:         "worker@example.com",
		EmailVerified: true,
	}, nil)

	client.EXPECT().GetUser(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *authv1.GetUserRequest, opts ...grpc.CallOption) (*authv1.GetUserResponse, error) {
			if req.GetUserId() != 7 {
				t.Errorf("user_id = %d, want 7", req.GetUserId())
			}

			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Errorf("missing outgoing metadata")
			} else {
				if got := md.Get("x-user-id"); len(got) != 1 || got[0] != "7" {
					t.Errorf("x-user-id metadata = %v, want [7]", got)
				}
				if got := md.Get("x-user-email-verified"); len(got) != 1 || got[0] != "true" {
					t.Errorf("x-user-email-verified metadata = %v, want [true]", got)
				}
			}
			return &authv1.GetUserResponse{User: sampleUser(true)}, nil
		},
	)

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodGet, "/v1/me", "", "access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestUpdateMeSendsOnlyProvidedFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockAuthServiceClient(ctrl)

	client.EXPECT().ValidateToken(gomock.Any(), gomock.Any()).Return(&authv1.ValidateTokenResponse{
		Valid:         true,
		UserId:        7,
		Role:          commonv1.UserRole_USER_ROLE_WORKER,
		Email:         "worker@example.com",
		EmailVerified: true,
	}, nil)

	client.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *authv1.UpdateProfileRequest, opts ...grpc.CallOption) (*authv1.UpdateProfileResponse, error) {
			if req.FirstName == nil || req.GetFirstName() != "Ivan" {
				t.Errorf("first_name = %v, want Ivan", req.FirstName)
			}
			if req.Phone != nil || req.AvatarUrl != nil || req.LastName != nil {
				t.Errorf("unexpected optional fields: %+v", req)
			}
			return &authv1.UpdateProfileResponse{User: sampleUser(true)}, nil
		},
	)

	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPatch, "/v1/me", `{"first_name":"Ivan"}`, "access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func newTestRouter(client authv1.AuthServiceClient) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestID())
	cfg := testGatewayConfig()
	RegisterRoutes(router.Group("/v1"), cfg, client, middleware.Auth(client, cfg), middleware.CSRF(cfg), middleware.RequireVerified())
	return router
}

func doJSON(router *gin.Engine, method string, path string, body string, accessToken string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.AddCookie(&http.Cookie{Name: "berusmenu_access", Value: accessToken})
		if method == http.MethodPost || method == http.MethodPatch || method == http.MethodPut || method == http.MethodDelete {
			req.AddCookie(&http.Cookie{Name: "berusmenu_csrf", Value: "csrf"})
			req.Header.Set("X-CSRF-Token", "csrf")
		}
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func cookieValue(rec *httptest.ResponseRecorder, name string) string {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func maxAgeForCookie(rec *httptest.ResponseRecorder, name string) int {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name {
			return cookie.MaxAge
		}
	}
	return 0
}

func testGatewayConfig() config.Config {
	return config.Config{
		AccessCookieName:  "berusmenu_access",
		RefreshCookieName: "berusmenu_refresh",
		CSRFCookieName:    "berusmenu_csrf",
		CSRFHeaderName:    "X-CSRF-Token",
		RefreshCookieTTL:  30 * 24 * time.Hour,
	}
}

func sampleAuthPayload(verified bool) *authv1.AuthPayload {
	return &authv1.AuthPayload{
		Tokens: &authv1.AuthTokens{
			AccessToken:     "access",
			RefreshToken:    "refresh",
			AccessExpiresAt: timestamppb.Now(),
		},
		User: sampleUser(verified),
	}
}

func sampleUser(verified bool) *authv1.User {
	user := &authv1.User{
		Id:        7,
		Email:     "worker@example.com",
		Role:      commonv1.UserRole_USER_ROLE_WORKER,
		LastName:  "User",
		FirstName: "Worker",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}
	if verified {
		user.EmailVerifiedAt = timestamppb.Now()
	}
	return user
}
