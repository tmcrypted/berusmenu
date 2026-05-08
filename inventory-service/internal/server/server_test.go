package server

import (
	"context"
	"testing"
	"time"

	"github.com/tmcrypted/berusmenu/inventory-service/internal/store"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	inventoryv1 "github.com/tmcrypted/berusmenu/proto/gen/go/inventory/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestCreatePvzRequiresIdentityMetadata(t *testing.T) {
	srv := New(Dependencies{Pvz: &serverFakeRepo{}, ActiveVacancies: serverFakeChecker{}})

	_, err := srv.CreatePvz(context.Background(), &inventoryv1.CreatePvzRequest{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("code = %s, want unauthenticated", status.Code(err))
	}
}

func TestCreatePvzMapsProtoAndMetadata(t *testing.T) {
	repo := &serverFakeRepo{}
	srv := New(Dependencies{Pvz: repo, ActiveVacancies: serverFakeChecker{}})
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-user-id", "7", "x-user-role", "owner"))

	res, err := srv.CreatePvz(ctx, &inventoryv1.CreatePvzRequest{
		City:         "Moscow",
		Address:      "Tverskaya 1",
		CompanyName:  "Beru",
		ContactPhone: "+79990001122",
	})
	if err != nil {
		t.Fatalf("CreatePvz returned error: %v", err)
	}
	if res.GetPvz().GetOwnerId() != 7 || res.GetPvz().GetStatus() != commonv1.PvzStatus_PVZ_STATUS_PENDING {
		t.Fatalf("pvz = %+v", res.GetPvz())
	}
}

func TestGetPvzMapsNotFound(t *testing.T) {
	srv := New(Dependencies{Pvz: &serverFakeRepo{err: store.ErrNotFound}, ActiveVacancies: serverFakeChecker{}})
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-user-id", "7", "x-user-role", "owner"))

	_, err := srv.GetPvz(ctx, &inventoryv1.GetPvzRequest{Id: 10})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("code = %s, want not found; err=%v", status.Code(err), err)
	}
}

type serverFakeRepo struct {
	err error
}

func (f *serverFakeRepo) CreatePvzWithEvent(ctx context.Context, pvz store.Pvz, actorUserID int64) (store.Pvz, error) {
	if f.err != nil {
		return store.Pvz{}, f.err
	}
	pvz.ID = 1
	pvz.Status = commonv1.PvzStatus_PVZ_STATUS_PENDING
	pvz.CreatedAt = time.Now()
	pvz.UpdatedAt = pvz.CreatedAt
	return pvz, nil
}

func (f *serverFakeRepo) GetPvz(ctx context.Context, id int64) (store.Pvz, error) {
	if f.err != nil {
		return store.Pvz{}, f.err
	}
	now := time.Now()
	return store.Pvz{ID: id, OwnerID: 7, Status: commonv1.PvzStatus_PVZ_STATUS_APPROVED, CreatedAt: now, UpdatedAt: now}, nil
}

func (f *serverFakeRepo) ListOwnerPvz(ctx context.Context, ownerID int64, limit uint64, offset uint64) ([]store.Pvz, error) {
	return []store.Pvz{}, f.err
}

func (f *serverFakeRepo) ListPendingPvz(ctx context.Context, limit uint64, offset uint64) ([]store.Pvz, error) {
	return []store.Pvz{}, f.err
}

func (f *serverFakeRepo) UpdatePvzWithEvent(ctx context.Context, id int64, patch store.PvzPatch, actorUserID int64) (store.Pvz, error) {
	if f.err != nil {
		return store.Pvz{}, f.err
	}
	now := time.Now()
	return store.Pvz{ID: id, OwnerID: actorUserID, Status: commonv1.PvzStatus_PVZ_STATUS_PENDING, CreatedAt: now, UpdatedAt: now}, nil
}

func (f *serverFakeRepo) ModeratePvzWithEvent(ctx context.Context, id int64, pvzStatus commonv1.PvzStatus, actorUserID int64, reason *string, action commonv1.PvzModerationAction, comment *string) (store.Pvz, error) {
	if f.err != nil {
		return store.Pvz{}, f.err
	}
	now := time.Now()
	return store.Pvz{ID: id, OwnerID: 7, Status: pvzStatus, CreatedAt: now, UpdatedAt: now}, nil
}

func (f *serverFakeRepo) DeletePendingAndBlockOwnerPvz(ctx context.Context, ownerID int64, actorUserID int64, comment *string) (int64, int64, error) {
	return 1, 2, f.err
}

type serverFakeChecker struct{}

func (serverFakeChecker) HasActiveVacancyForPvz(ctx context.Context, pvzID int64) (bool, error) {
	return false, nil
}
