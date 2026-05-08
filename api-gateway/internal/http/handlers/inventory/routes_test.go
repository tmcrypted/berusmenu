package inventory

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
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	inventoryv1 "github.com/tmcrypted/berusmenu/proto/gen/go/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreatePvzForwardsRequestAndIdentity(t *testing.T) {
	inventoryClient := &fakeInventoryClient{
		createPvz: func(ctx context.Context, req *inventoryv1.CreatePvzRequest, opts ...grpc.CallOption) (*inventoryv1.CreatePvzResponse, error) {
			if req.GetCity() != "Moscow" || req.GetCompanyName() != "Beru" {
				t.Fatalf("request = %+v", req)
			}
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok || len(md.Get("x-user-id")) != 1 || md.Get("x-user-id")[0] != "7" {
				t.Fatalf("missing identity metadata: %v", md)
			}
			return &inventoryv1.CreatePvzResponse{Pvz: sampleGatewayPvz()}, nil
		},
	}
	router := newInventoryTestRouter(inventoryClient, true)

	rec := doInventoryJSON(router, http.MethodPost, "/v1/pvz", `{
		"city":"Moscow",
		"address":"Tverskaya 1",
		"company_name":"Beru",
		"contact_phone":"+79990001122"
	}`, "access")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"company_name":"Beru"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestInventoryRoutesRequireVerifiedUser(t *testing.T) {
	router := newInventoryTestRouter(&fakeInventoryClient{}, false)

	rec := doInventoryJSON(router, http.MethodGet, "/v1/pvz", "", "access")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestPatchPvzMapsGRPCPrecondition(t *testing.T) {
	inventoryClient := &fakeInventoryClient{
		updatePvz: func(ctx context.Context, req *inventoryv1.UpdatePvzRequest, opts ...grpc.CallOption) (*inventoryv1.UpdatePvzResponse, error) {
			if req.GetId() != 10 || req.City == nil || req.GetCity() != "Kazan" {
				t.Fatalf("request = %+v", req)
			}
			return nil, status.Error(codes.FailedPrecondition, "pvz has active vacancies")
		},
	}
	router := newInventoryTestRouter(inventoryClient, true)

	rec := doInventoryJSON(router, http.MethodPatch, "/v1/pvz/10", `{"city":"Kazan"}`, "access")

	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusPreconditionFailed, rec.Body.String())
	}
}

func TestRejectPvzRequiresReason(t *testing.T) {
	router := newInventoryTestRouter(&fakeInventoryClient{}, true)

	rec := doInventoryJSON(router, http.MethodPost, "/v1/moderation/pvz/10/reject", `{}`, "access")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestListPendingPvzParsesPagination(t *testing.T) {
	inventoryClient := &fakeInventoryClient{
		listPendingPvz: func(ctx context.Context, req *inventoryv1.ListPendingPvzRequest, opts ...grpc.CallOption) (*inventoryv1.ListPendingPvzResponse, error) {
			if req.GetPageSize() != 10 || req.GetPageToken() != "20" {
				t.Fatalf("request = %+v", req)
			}
			return &inventoryv1.ListPendingPvzResponse{}, nil
		},
	}
	router := newInventoryTestRouter(inventoryClient, true)

	rec := doInventoryJSON(router, http.MethodGet, "/v1/moderation/pvz?page_size=10&page_token=20", "", "access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestBlockUserAndDeletePvzCallsAuthThenInventory(t *testing.T) {
	authClient := &fakeAuthClient{verified: true, role: commonv1.UserRole_USER_ROLE_MODERATOR}
	inventoryClient := &fakeInventoryClient{
		deleteOwnerPvzRequests: func(ctx context.Context, req *inventoryv1.DeleteOwnerPvzRequestsRequest, opts ...grpc.CallOption) (*inventoryv1.DeleteOwnerPvzRequestsResponse, error) {
			if req.GetOwnerId() != 42 || req.GetComment() != "spam cleanup" {
				t.Fatalf("inventory request = %+v", req)
			}
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok || md.Get("x-user-role")[0] != "moderator" {
				t.Fatalf("missing moderator metadata: %v", md)
			}
			return &inventoryv1.DeleteOwnerPvzRequestsResponse{DeletedCount: 3, BlockedCount: 2}, nil
		},
	}
	router := newInventoryTestRouterWithAuth(inventoryClient, authClient)

	rec := doInventoryJSON(router, http.MethodPost, "/v1/moderation/users/42/block-and-delete-pvz", `{
		"reason":"spam",
		"block_duration_minutes":60,
		"comment":"spam cleanup"
	}`, "access")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"deleted_count":"3"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"blocked_count":"2"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
	if authClient.blockedUserID != 42 || authClient.blockReason != "spam" || authClient.blockedUntil.IsZero() {
		t.Fatalf("auth block call = id %d reason %q until %s", authClient.blockedUserID, authClient.blockReason, authClient.blockedUntil)
	}
}

func TestBlockUserAndDeletePvzRequiresReason(t *testing.T) {
	router := newInventoryTestRouterWithAuth(&fakeInventoryClient{}, &fakeAuthClient{verified: true, role: commonv1.UserRole_USER_ROLE_MODERATOR})

	rec := doInventoryJSON(router, http.MethodPost, "/v1/moderation/users/42/block-and-delete-pvz", `{}`, "access")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func newInventoryTestRouter(inventoryClient inventoryv1.InventoryServiceClient, verified bool) *gin.Engine {
	return newInventoryTestRouterWithAuth(inventoryClient, &fakeAuthClient{verified: verified, role: commonv1.UserRole_USER_ROLE_OWNER})
}

func newInventoryTestRouterWithAuth(inventoryClient inventoryv1.InventoryServiceClient, authClient *fakeAuthClient) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestID())
	cfg := config.Config{
		AccessCookieName: "berusmenu_access",
		CSRFCookieName:   "berusmenu_csrf",
		CSRFHeaderName:   "X-CSRF-Token",
	}
	RegisterRoutes(router.Group("/v1"), inventoryClient, authClient, middleware.Auth(authClient, cfg), middleware.CSRF(cfg), middleware.RequireVerified())
	return router
}

func doInventoryJSON(router *gin.Engine, method string, path string, body string, accessToken string) *httptest.ResponseRecorder {
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

func sampleGatewayPvz() *inventoryv1.Pvz {
	now := timestamppb.New(time.Now())
	return &inventoryv1.Pvz{
		Id:           10,
		OwnerId:      7,
		Status:       commonv1.PvzStatus_PVZ_STATUS_PENDING,
		City:         "Moscow",
		Address:      "Tverskaya 1",
		CompanyName:  "Beru",
		ContactPhone: "+79990001122",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

type fakeAuthClient struct {
	authv1.AuthServiceClient
	verified      bool
	role          commonv1.UserRole
	blockedUserID int64
	blockReason   string
	blockedUntil  time.Time
}

func (f *fakeAuthClient) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest, opts ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
	role := f.role
	if role == commonv1.UserRole_USER_ROLE_UNSPECIFIED {
		role = commonv1.UserRole_USER_ROLE_OWNER
	}
	return &authv1.ValidateTokenResponse{
		Valid:         true,
		UserId:        7,
		Role:          role,
		Email:         "owner@example.com",
		EmailVerified: f.verified,
	}, nil
}

func (f *fakeAuthClient) BlockUser(ctx context.Context, req *authv1.BlockUserRequest, opts ...grpc.CallOption) (*authv1.BlockUserResponse, error) {
	f.blockedUserID = req.GetUserId()
	f.blockReason = req.GetReason()
	f.blockedUntil = req.GetBlockedUntil().AsTime()
	return &authv1.BlockUserResponse{User: &authv1.User{Id: req.GetUserId()}}, nil
}

type fakeInventoryClient struct {
	inventoryv1.InventoryServiceClient
	createPvz              func(context.Context, *inventoryv1.CreatePvzRequest, ...grpc.CallOption) (*inventoryv1.CreatePvzResponse, error)
	listOwnerPvz           func(context.Context, *inventoryv1.ListOwnerPvzRequest, ...grpc.CallOption) (*inventoryv1.ListOwnerPvzResponse, error)
	getPvz                 func(context.Context, *inventoryv1.GetPvzRequest, ...grpc.CallOption) (*inventoryv1.GetPvzResponse, error)
	updatePvz              func(context.Context, *inventoryv1.UpdatePvzRequest, ...grpc.CallOption) (*inventoryv1.UpdatePvzResponse, error)
	listPendingPvz         func(context.Context, *inventoryv1.ListPendingPvzRequest, ...grpc.CallOption) (*inventoryv1.ListPendingPvzResponse, error)
	deleteOwnerPvzRequests func(context.Context, *inventoryv1.DeleteOwnerPvzRequestsRequest, ...grpc.CallOption) (*inventoryv1.DeleteOwnerPvzRequestsResponse, error)
}

func (f *fakeInventoryClient) CreatePvz(ctx context.Context, req *inventoryv1.CreatePvzRequest, opts ...grpc.CallOption) (*inventoryv1.CreatePvzResponse, error) {
	if f.createPvz != nil {
		return f.createPvz(ctx, req, opts...)
	}
	return &inventoryv1.CreatePvzResponse{Pvz: sampleGatewayPvz()}, nil
}

func (f *fakeInventoryClient) GetPvz(ctx context.Context, req *inventoryv1.GetPvzRequest, opts ...grpc.CallOption) (*inventoryv1.GetPvzResponse, error) {
	if f.getPvz != nil {
		return f.getPvz(ctx, req, opts...)
	}
	return &inventoryv1.GetPvzResponse{Pvz: sampleGatewayPvz()}, nil
}

func (f *fakeInventoryClient) ListOwnerPvz(ctx context.Context, req *inventoryv1.ListOwnerPvzRequest, opts ...grpc.CallOption) (*inventoryv1.ListOwnerPvzResponse, error) {
	if f.listOwnerPvz != nil {
		return f.listOwnerPvz(ctx, req, opts...)
	}
	return &inventoryv1.ListOwnerPvzResponse{}, nil
}

func (f *fakeInventoryClient) UpdatePvz(ctx context.Context, req *inventoryv1.UpdatePvzRequest, opts ...grpc.CallOption) (*inventoryv1.UpdatePvzResponse, error) {
	if f.updatePvz != nil {
		return f.updatePvz(ctx, req, opts...)
	}
	return &inventoryv1.UpdatePvzResponse{Pvz: sampleGatewayPvz()}, nil
}

func (f *fakeInventoryClient) CheckPvzForVacancy(ctx context.Context, req *inventoryv1.CheckPvzForVacancyRequest, opts ...grpc.CallOption) (*inventoryv1.CheckPvzForVacancyResponse, error) {
	return &inventoryv1.CheckPvzForVacancyResponse{}, nil
}

func (f *fakeInventoryClient) ListPendingPvz(ctx context.Context, req *inventoryv1.ListPendingPvzRequest, opts ...grpc.CallOption) (*inventoryv1.ListPendingPvzResponse, error) {
	if f.listPendingPvz != nil {
		return f.listPendingPvz(ctx, req, opts...)
	}
	return &inventoryv1.ListPendingPvzResponse{}, nil
}

func (f *fakeInventoryClient) ApprovePvz(ctx context.Context, req *inventoryv1.ApprovePvzRequest, opts ...grpc.CallOption) (*inventoryv1.ApprovePvzResponse, error) {
	return &inventoryv1.ApprovePvzResponse{Pvz: sampleGatewayPvz()}, nil
}

func (f *fakeInventoryClient) RejectPvz(ctx context.Context, req *inventoryv1.RejectPvzRequest, opts ...grpc.CallOption) (*inventoryv1.RejectPvzResponse, error) {
	return &inventoryv1.RejectPvzResponse{Pvz: sampleGatewayPvz()}, nil
}

func (f *fakeInventoryClient) BlockPvz(ctx context.Context, req *inventoryv1.BlockPvzRequest, opts ...grpc.CallOption) (*inventoryv1.BlockPvzResponse, error) {
	return &inventoryv1.BlockPvzResponse{Pvz: sampleGatewayPvz()}, nil
}

func (f *fakeInventoryClient) DeleteOwnerPvzRequests(ctx context.Context, req *inventoryv1.DeleteOwnerPvzRequestsRequest, opts ...grpc.CallOption) (*inventoryv1.DeleteOwnerPvzRequestsResponse, error) {
	if f.deleteOwnerPvzRequests != nil {
		return f.deleteOwnerPvzRequests(ctx, req, opts...)
	}
	return &inventoryv1.DeleteOwnerPvzRequestsResponse{}, nil
}
