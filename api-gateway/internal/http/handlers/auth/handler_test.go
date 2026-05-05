package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gitlab.com/berusmenu/api-gateway/internal/http/middleware"
	authv1 "gitlab.com/berusmenu/proto/gen/go/auth/v1"
	commonv1 "gitlab.com/berusmenu/proto/gen/go/common/v1"
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
	client := &authClientFake{
		registerResp: &authv1.RegisterResponse{Auth: sampleAuthPayload(false)},
	}
	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPost, "/v1/auth/register", `{
		"email":"owner@example.com",
		"password":"password123",
		"full_name":"Owner User",
		"role":"owner"
	}`, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if client.registerReq.GetRole() != commonv1.UserRole_USER_ROLE_OWNER {
		t.Fatalf("role = %s, want owner", client.registerReq.GetRole())
	}
	if client.registerReq.GetFullName() != "Owner User" {
		t.Fatalf("full_name = %q, want Owner User", client.registerReq.GetFullName())
	}
	if !strings.Contains(rec.Body.String(), `"access_token":"access"`) {
		t.Fatalf("response does not contain access token: %s", rec.Body.String())
	}
}

func TestRegisterValidationFailure(t *testing.T) {
	router := newTestRouter(&authClientFake{})

	rec := doJSON(router, http.MethodPost, "/v1/auth/register", `{
		"email":"bad",
		"password":"short",
		"full_name":"User"
	}`, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLoginWrongCredentials(t *testing.T) {
	client := &authClientFake{
		loginErr: status.Error(codes.Unauthenticated, "bad credentials"),
	}
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
	router := newTestRouter(&authClientFake{})

	rec := doJSON(router, http.MethodPost, "/v1/auth/resend-code", `{}`, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestResendVerificationCodeSuccess(t *testing.T) {
	client := &authClientFake{
		validateResp: &authv1.ValidateTokenResponse{
			Valid:         true,
			UserId:        7,
			Role:          commonv1.UserRole_USER_ROLE_WORKER,
			Email:         "worker@example.com",
			EmailVerified: false,
		},
		resendResp: &authv1.ResendVerificationCodeResponse{
			CodeExpiresAt:     timestamppb.Now(),
			ResendAvailableAt: timestamppb.Now(),
		},
	}
	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPost, "/v1/auth/resend-code", `{}`, "Bearer access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestVerifyEmailSuccess(t *testing.T) {
	client := &authClientFake{
		validateResp: &authv1.ValidateTokenResponse{
			Valid:         true,
			UserId:        7,
			Role:          commonv1.UserRole_USER_ROLE_WORKER,
			Email:         "worker@example.com",
			EmailVerified: false,
		},
		verifyResp: &authv1.VerifyEmailResponse{User: sampleUser(true)},
	}
	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPost, "/v1/auth/verify-email", `{"code":"123456"}`, "Bearer access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if client.verifyReq.GetCode() != "123456" {
		t.Fatalf("code = %q, want 123456", client.verifyReq.GetCode())
	}
}

func TestGetMeRequiresToken(t *testing.T) {
	router := newTestRouter(&authClientFake{})

	rec := doJSON(router, http.MethodGet, "/v1/me", "", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetMeRejectsUnverifiedUser(t *testing.T) {
	client := &authClientFake{
		validateResp: &authv1.ValidateTokenResponse{
			Valid:         true,
			UserId:        7,
			Role:          commonv1.UserRole_USER_ROLE_WORKER,
			Email:         "worker@example.com",
			EmailVerified: false,
		},
	}
	router := newTestRouter(client)

	rec := doJSON(router, http.MethodGet, "/v1/me", "", "Bearer access")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestGetMeSuccessUsesIdentityMetadata(t *testing.T) {
	client := &authClientFake{
		validateResp: &authv1.ValidateTokenResponse{
			Valid:         true,
			UserId:        7,
			Role:          commonv1.UserRole_USER_ROLE_WORKER,
			Email:         "worker@example.com",
			EmailVerified: true,
		},
		getUserResp: &authv1.GetUserResponse{User: sampleUser(true)},
	}
	router := newTestRouter(client)

	rec := doJSON(router, http.MethodGet, "/v1/me", "", "Bearer access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if client.getUserReq.GetUserId() != 7 {
		t.Fatalf("user_id = %d, want 7", client.getUserReq.GetUserId())
	}
	if got := client.outgoingMD.Get("x-user-id"); len(got) != 1 || got[0] != "7" {
		t.Fatalf("x-user-id metadata = %v, want [7]", got)
	}
	if got := client.outgoingMD.Get("x-user-email-verified"); len(got) != 1 || got[0] != "true" {
		t.Fatalf("x-user-email-verified metadata = %v, want [true]", got)
	}
}

func TestUpdateMeSendsOnlyProvidedFields(t *testing.T) {
	client := &authClientFake{
		validateResp: &authv1.ValidateTokenResponse{
			Valid:         true,
			UserId:        7,
			Role:          commonv1.UserRole_USER_ROLE_WORKER,
			Email:         "worker@example.com",
			EmailVerified: true,
		},
		updateResp: &authv1.UpdateProfileResponse{User: sampleUser(true)},
	}
	router := newTestRouter(client)

	rec := doJSON(router, http.MethodPatch, "/v1/me", `{"full_name":"New Name"}`, "Bearer access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if client.updateReq.FullName == nil || client.updateReq.GetFullName() != "New Name" {
		t.Fatalf("full_name = %v, want New Name", client.updateReq.FullName)
	}
	if client.updateReq.Phone != nil || client.updateReq.AvatarUrl != nil {
		t.Fatalf("unexpected optional fields: %+v", client.updateReq)
	}
}

func newTestRouter(client authv1.AuthServiceClient) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestID())
	RegisterRoutes(router.Group("/v1"), client, middleware.Auth(client), middleware.RequireVerified())
	return router
}

func doJSON(router *gin.Engine, method string, path string, body string, authHeader string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
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
		FullName:  "Worker User",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}
	if verified {
		user.EmailVerifiedAt = timestamppb.Now()
	}
	return user
}

type authClientFake struct {
	registerReq  *authv1.RegisterRequest
	registerResp *authv1.RegisterResponse
	registerErr  error

	loginResp *authv1.LoginResponse
	loginErr  error

	refreshResp *authv1.RefreshResponse
	refreshErr  error

	logoutErr error

	validateResp *authv1.ValidateTokenResponse
	validateErr  error

	getUserReq  *authv1.GetUserRequest
	getUserResp *authv1.GetUserResponse
	getUserErr  error

	updateReq  *authv1.UpdateProfileRequest
	updateResp *authv1.UpdateProfileResponse
	updateErr  error

	resendResp *authv1.ResendVerificationCodeResponse
	resendErr  error

	verifyReq  *authv1.VerifyEmailRequest
	verifyResp *authv1.VerifyEmailResponse
	verifyErr  error

	outgoingMD metadata.MD
}

func (f *authClientFake) Register(ctx context.Context, req *authv1.RegisterRequest, opts ...grpc.CallOption) (*authv1.RegisterResponse, error) {
	f.registerReq = req
	if f.registerErr != nil {
		return nil, f.registerErr
	}
	return f.registerResp, nil
}

func (f *authClientFake) Login(ctx context.Context, req *authv1.LoginRequest, opts ...grpc.CallOption) (*authv1.LoginResponse, error) {
	if f.loginErr != nil {
		return nil, f.loginErr
	}
	if f.loginResp != nil {
		return f.loginResp, nil
	}
	return &authv1.LoginResponse{Auth: sampleAuthPayload(true)}, nil
}

func (f *authClientFake) Refresh(ctx context.Context, req *authv1.RefreshRequest, opts ...grpc.CallOption) (*authv1.RefreshResponse, error) {
	if f.refreshErr != nil {
		return nil, f.refreshErr
	}
	if f.refreshResp != nil {
		return f.refreshResp, nil
	}
	return &authv1.RefreshResponse{Auth: sampleAuthPayload(true)}, nil
}

func (f *authClientFake) Logout(ctx context.Context, req *authv1.LogoutRequest, opts ...grpc.CallOption) (*authv1.LogoutResponse, error) {
	if f.logoutErr != nil {
		return nil, f.logoutErr
	}
	return &authv1.LogoutResponse{}, nil
}

func (f *authClientFake) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest, opts ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
	if f.validateErr != nil {
		return nil, f.validateErr
	}
	if f.validateResp != nil {
		return f.validateResp, nil
	}
	return &authv1.ValidateTokenResponse{Valid: false}, nil
}

func (f *authClientFake) GetUser(ctx context.Context, req *authv1.GetUserRequest, opts ...grpc.CallOption) (*authv1.GetUserResponse, error) {
	f.getUserReq = req
	f.outgoingMD, _ = metadata.FromOutgoingContext(ctx)
	if f.getUserErr != nil {
		return nil, f.getUserErr
	}
	return f.getUserResp, nil
}

func (f *authClientFake) UpdateProfile(ctx context.Context, req *authv1.UpdateProfileRequest, opts ...grpc.CallOption) (*authv1.UpdateProfileResponse, error) {
	f.updateReq = req
	f.outgoingMD, _ = metadata.FromOutgoingContext(ctx)
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return f.updateResp, nil
}

func (f *authClientFake) ResendVerificationCode(ctx context.Context, req *authv1.ResendVerificationCodeRequest, opts ...grpc.CallOption) (*authv1.ResendVerificationCodeResponse, error) {
	if f.resendErr != nil {
		return nil, f.resendErr
	}
	if f.resendResp != nil {
		return f.resendResp, nil
	}
	return &authv1.ResendVerificationCodeResponse{}, nil
}

func (f *authClientFake) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest, opts ...grpc.CallOption) (*authv1.VerifyEmailResponse, error) {
	f.verifyReq = req
	if f.verifyErr != nil {
		return nil, f.verifyErr
	}
	return f.verifyResp, nil
}
