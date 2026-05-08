package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tmcrypted/berusmenu/inventory-service/internal/dto"
	"github.com/tmcrypted/berusmenu/inventory-service/internal/store"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
)

func TestCreatePvzRequiresOwnerAndCreatesPending(t *testing.T) {
	repo := newFakeRepo()
	svc := New(repo, fakeChecker{})

	pvz, err := svc.CreatePvz(context.Background(), dto.CreatePvzInput{
		Actor:        ownerActor(),
		City:         " Moscow ",
		Address:      " Tverskaya 1 ",
		CompanyName:  " Beru ",
		ContactPhone: " +79990001122 ",
	})
	if err != nil {
		t.Fatalf("CreatePvz returned error: %v", err)
	}
	if pvz.Status != commonv1.PvzStatus_PVZ_STATUS_PENDING {
		t.Fatalf("status = %s, want pending", pvz.Status)
	}
	if repo.submittedEvents != 1 {
		t.Fatalf("submitted events = %d, want 1", repo.submittedEvents)
	}

	_, err = svc.CreatePvz(context.Background(), dto.CreatePvzInput{
		Actor:        workerActor(),
		City:         "Moscow",
		Address:      "Tverskaya 1",
		CompanyName:  "Beru",
		ContactPhone: "+79990001122",
	})
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("worker CreatePvz error = %v, want permission denied", err)
	}
}

func TestUpdatePvzChecksOwnerAndActiveVacancy(t *testing.T) {
	repo := newFakeRepo()
	repo.items[10] = samplePvz(10, 7, commonv1.PvzStatus_PVZ_STATUS_APPROVED)

	svc := New(repo, fakeChecker{hasActive: true})
	city := "Kazan"
	_, err := svc.UpdatePvz(context.Background(), dto.UpdatePvzInput{
		Actor: ownerActor(),
		ID:    10,
		City:  &city,
	})
	if !errors.Is(err, ErrFailedPrecondition) {
		t.Fatalf("active vacancy error = %v, want failed precondition", err)
	}
	if repo.items[10].City != "Moscow" {
		t.Fatalf("pvz mutated on denied update")
	}

	svc = New(repo, fakeChecker{err: errors.New("matching down")})
	_, err = svc.UpdatePvz(context.Background(), dto.UpdatePvzInput{
		Actor: ownerActor(),
		ID:    10,
		City:  &city,
	})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("checker error = %v, want unavailable", err)
	}

	svc = New(repo, fakeChecker{})
	updated, err := svc.UpdatePvz(context.Background(), dto.UpdatePvzInput{
		Actor: ownerActor(),
		ID:    10,
		City:  &city,
	})
	if err != nil {
		t.Fatalf("UpdatePvz returned error: %v", err)
	}
	if updated.City != "Kazan" || updated.Status != commonv1.PvzStatus_PVZ_STATUS_PENDING {
		t.Fatalf("updated pvz = %+v, want city Kazan and pending", updated)
	}
	if updated.ModeratedBy != nil || updated.ModeratedAt != nil || updated.RejectionReason != nil {
		t.Fatalf("moderation fields were not cleared: %+v", updated)
	}
}

func TestModerationRequiresModerator(t *testing.T) {
	repo := newFakeRepo()
	repo.items[10] = samplePvz(10, 7, commonv1.PvzStatus_PVZ_STATUS_PENDING)
	svc := New(repo, fakeChecker{})

	_, err := svc.ApprovePvz(context.Background(), dto.ModeratePvzInput{Actor: ownerActor(), PvzID: 10})
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("owner approve error = %v, want permission denied", err)
	}

	approved, err := svc.ApprovePvz(context.Background(), dto.ModeratePvzInput{Actor: moderatorActor(), PvzID: 10})
	if err != nil {
		t.Fatalf("ApprovePvz returned error: %v", err)
	}
	if approved.Status != commonv1.PvzStatus_PVZ_STATUS_APPROVED || repo.approvedEvents != 1 {
		t.Fatalf("approved pvz=%+v approved_events=%d", approved, repo.approvedEvents)
	}
}

func TestRejectRequiresReason(t *testing.T) {
	repo := newFakeRepo()
	repo.items[10] = samplePvz(10, 7, commonv1.PvzStatus_PVZ_STATUS_PENDING)
	svc := New(repo, fakeChecker{})

	_, err := svc.RejectPvz(context.Background(), dto.ModeratePvzInput{Actor: moderatorActor(), PvzID: 10})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("reject without reason error = %v, want invalid argument", err)
	}

	reason := "bad documents"
	rejected, err := svc.RejectPvz(context.Background(), dto.ModeratePvzInput{Actor: moderatorActor(), PvzID: 10, Reason: &reason})
	if err != nil {
		t.Fatalf("RejectPvz returned error: %v", err)
	}
	if rejected.RejectionReason == nil || *rejected.RejectionReason != reason {
		t.Fatalf("rejection reason = %v, want %q", rejected.RejectionReason, reason)
	}
}

func TestCheckPvzForVacancyRequiresApprovedOwnPvz(t *testing.T) {
	repo := newFakeRepo()
	repo.items[10] = samplePvz(10, 7, commonv1.PvzStatus_PVZ_STATUS_PENDING)
	svc := New(repo, fakeChecker{})

	_, err := svc.CheckPvzForVacancy(context.Background(), dto.CheckPvzForVacancyInput{Actor: ownerActor(), PvzID: 10})
	if !errors.Is(err, ErrFailedPrecondition) {
		t.Fatalf("pending pvz error = %v, want failed precondition", err)
	}

	repo.items[10] = samplePvz(10, 7, commonv1.PvzStatus_PVZ_STATUS_APPROVED)
	pvz, err := svc.CheckPvzForVacancy(context.Background(), dto.CheckPvzForVacancyInput{Actor: ownerActor(), PvzID: 10})
	if err != nil {
		t.Fatalf("CheckPvzForVacancy returned error: %v", err)
	}
	if pvz.ID != 10 {
		t.Fatalf("pvz id = %d, want 10", pvz.ID)
	}
}

type fakeRepo struct {
	items           map[int64]store.Pvz
	nextID          int64
	submittedEvents int
	approvedEvents  int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{items: make(map[int64]store.Pvz), nextID: 1}
}

func (f *fakeRepo) CreatePvzWithEvent(ctx context.Context, pvz store.Pvz, actorUserID int64) (store.Pvz, error) {
	pvz.ID = f.nextID
	f.nextID++
	pvz.Status = commonv1.PvzStatus_PVZ_STATUS_PENDING
	pvz.CreatedAt = time.Now()
	pvz.UpdatedAt = pvz.CreatedAt
	f.items[pvz.ID] = pvz
	f.submittedEvents++
	return pvz, nil
}

func (f *fakeRepo) GetPvz(ctx context.Context, id int64) (store.Pvz, error) {
	pvz, ok := f.items[id]
	if !ok {
		return store.Pvz{}, store.ErrNotFound
	}
	return pvz, nil
}

func (f *fakeRepo) ListOwnerPvz(ctx context.Context, ownerID int64, limit uint64, offset uint64) ([]store.Pvz, error) {
	return nil, nil
}

func (f *fakeRepo) ListPendingPvz(ctx context.Context, limit uint64, offset uint64) ([]store.Pvz, error) {
	return nil, nil
}

func (f *fakeRepo) UpdatePvzWithEvent(ctx context.Context, id int64, patch store.PvzPatch, actorUserID int64) (store.Pvz, error) {
	pvz, ok := f.items[id]
	if !ok {
		return store.Pvz{}, store.ErrNotFound
	}
	if patch.City != nil {
		pvz.City = *patch.City
	}
	pvz.Status = commonv1.PvzStatus_PVZ_STATUS_PENDING
	pvz.RejectionReason = nil
	pvz.ModeratedBy = nil
	pvz.ModeratedAt = nil
	f.items[id] = pvz
	f.submittedEvents++
	return pvz, nil
}

func (f *fakeRepo) ModeratePvzWithEvent(ctx context.Context, id int64, status commonv1.PvzStatus, actorUserID int64, reason *string, action commonv1.PvzModerationAction, comment *string) (store.Pvz, error) {
	pvz, ok := f.items[id]
	if !ok {
		return store.Pvz{}, store.ErrNotFound
	}
	pvz.Status = status
	pvz.RejectionReason = reason
	pvz.ModeratedBy = &actorUserID
	now := time.Now()
	pvz.ModeratedAt = &now
	f.items[id] = pvz
	if action == commonv1.PvzModerationAction_PVZ_MODERATION_ACTION_APPROVED {
		f.approvedEvents++
	}
	return pvz, nil
}

type fakeChecker struct {
	hasActive bool
	err       error
}

func (f fakeChecker) HasActiveVacancyForPvz(ctx context.Context, pvzID int64) (bool, error) {
	return f.hasActive, f.err
}

func samplePvz(id int64, ownerID int64, status commonv1.PvzStatus) store.Pvz {
	return store.Pvz{
		ID:           id,
		OwnerID:      ownerID,
		Status:       status,
		City:         "Moscow",
		Address:      "Tverskaya 1",
		CompanyName:  "Beru",
		ContactPhone: "+79990001122",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func ownerActor() dto.Actor {
	return dto.Actor{UserID: 7, Role: commonv1.UserRole_USER_ROLE_OWNER}
}

func workerActor() dto.Actor {
	return dto.Actor{UserID: 8, Role: commonv1.UserRole_USER_ROLE_WORKER}
}

func moderatorActor() dto.Actor {
	return dto.Actor{UserID: 9, Role: commonv1.UserRole_USER_ROLE_MODERATOR}
}
