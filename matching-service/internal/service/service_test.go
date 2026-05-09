package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tmcrypted/berusmenu/matching-service/internal/dto"
	"github.com/tmcrypted/berusmenu/matching-service/internal/store"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
)

func TestCreateVacancyRejectsNonOneSlotMVP(t *testing.T) {
	svc := New(&fakeRepo{}, fakeInventory{}, fakeAuth{}, fakeEvents{})
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	_, err := svc.CreateVacancy(context.Background(), dto.CreateVacancyInput{
		Actor:        dto.Actor{UserID: 10, Role: commonv1.UserRole_USER_ROLE_OWNER},
		PvzID:        1,
		ShiftStartAt: now.Add(time.Hour),
		ShiftEndAt:   now.Add(9 * time.Hour),
		PayAmount:    3000,
		SlotsTotal:   2,
	})

	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}

func TestCreateResponseRejectsOwnVacancy(t *testing.T) {
	repo := &fakeRepo{
		vacancy: store.Vacancy{
			ID:           1,
			OwnerID:      10,
			Status:       commonv1.VacancyStatus_VACANCY_STATUS_PUBLISHED,
			ShiftStartAt: time.Now().Add(time.Hour),
			ShiftEndAt:   time.Now().Add(9 * time.Hour),
			PayAmount:    3000,
			SlotsTotal:   1,
		},
	}
	svc := New(repo, fakeInventory{}, fakeAuth{}, fakeEvents{})

	_, err := svc.CreateResponse(context.Background(), dto.CreateResponseInput{
		Actor:     dto.Actor{UserID: 10, Role: commonv1.UserRole_USER_ROLE_WORKER},
		VacancyID: 1,
	})

	if !errors.Is(err, ErrFailedPrecondition) {
		t.Fatalf("expected failed precondition, got %v", err)
	}
}

func TestCreateReviewRejectsRatingOutsideScale(t *testing.T) {
	repo := &fakeRepo{
		assignment: store.Assignment{
			ID:       1,
			OwnerID:  10,
			WorkerID: 20,
			Status:   commonv1.AssignmentStatus_ASSIGNMENT_STATUS_COMPLETED,
		},
	}
	svc := New(repo, fakeInventory{}, fakeAuth{}, fakeEvents{})

	_, err := svc.CreateReview(context.Background(), dto.CreateReviewInput{
		Actor:        dto.Actor{UserID: 10, Role: commonv1.UserRole_USER_ROLE_OWNER},
		AssignmentID: 1,
		Rating:       6,
	})

	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}

func TestModeratorAcceptRequiresReason(t *testing.T) {
	repo := &fakeRepo{
		vacancy: store.Vacancy{
			ID:      1,
			OwnerID: 10,
			Status:  commonv1.VacancyStatus_VACANCY_STATUS_PUBLISHED,
		},
		response: store.Response{
			ID:        2,
			VacancyID: 1,
			WorkerID:  20,
			Status:    commonv1.ResponseStatus_RESPONSE_STATUS_PENDING,
		},
	}
	svc := New(repo, fakeInventory{}, fakeAuth{}, fakeEvents{})

	_, err := svc.AcceptResponse(context.Background(), dto.AcceptResponseInput{
		Actor:      dto.Actor{UserID: 99, Role: commonv1.UserRole_USER_ROLE_MODERATOR},
		ResponseID: 2,
	})

	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}

func TestModeratorCancelPassesAdminEventToRepository(t *testing.T) {
	repo := &fakeRepo{
		assignment: store.Assignment{
			ID:       1,
			OwnerID:  10,
			WorkerID: 20,
			Status:   commonv1.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE,
		},
	}
	svc := New(repo, fakeInventory{}, fakeAuth{}, fakeEvents{})
	reason := "dispute resolved"

	_, err := svc.CancelAssignment(context.Background(), dto.CancelAssignmentInput{
		Actor:        dto.Actor{UserID: 99, Role: commonv1.UserRole_USER_ROLE_MODERATOR},
		AssignmentID: 1,
		AdminReason:  &reason,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.finalizeAdmin == nil {
		t.Fatal("expected admin event to be passed to repository")
	}
	if repo.finalizeAdmin.Reason != reason {
		t.Fatalf("expected reason %q, got %q", reason, repo.finalizeAdmin.Reason)
	}
}

func TestCreateReviewReplyRejectsNonParticipantTarget(t *testing.T) {
	repo := &fakeRepo{
		review: store.Review{
			ID:           1,
			AssignmentID: 2,
		},
		assignment: store.Assignment{
			ID:       2,
			OwnerID:  10,
			WorkerID: 20,
		},
	}
	svc := New(repo, fakeInventory{}, fakeAuth{}, fakeEvents{})

	_, err := svc.CreateReviewReply(context.Background(), dto.CreateReviewReplyInput{
		Actor:        dto.Actor{UserID: 10, Role: commonv1.UserRole_USER_ROLE_OWNER},
		ReviewID:     1,
		TargetUserID: 999,
		Body:         "hello",
	})

	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}

func TestGetUserRatingReturnsZeroAggregateWhenMissing(t *testing.T) {
	svc := New(&fakeRepo{}, fakeInventory{}, fakeAuth{}, fakeEvents{})

	rating, err := svc.GetUserRating(context.Background(), dto.GetUserRatingInput{
		Actor:  dto.Actor{UserID: 10, Role: commonv1.UserRole_USER_ROLE_WORKER},
		UserID: 20,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rating.UserID != 20 || rating.AvgRating != 0 || rating.ReviewsCount != 0 {
		t.Fatalf("unexpected rating aggregate: %+v", rating)
	}
}

type fakeInventory struct{}

func (fakeInventory) CheckPvzForVacancy(ctx context.Context, pvzID int64) (dto.PvzSnapshot, error) {
	return dto.PvzSnapshot{PvzID: pvzID, OwnerID: 10, CompanyName: "PVZ", Address: "Address"}, nil
}

type fakeAuth struct{}

func (fakeAuth) GetUserByID(ctx context.Context, userID int64) (dto.UserSnapshot, error) {
	return dto.UserSnapshot{UserID: userID, Email: "user@example.com", FullName: "User", Role: commonv1.UserRole_USER_ROLE_OWNER}, nil
}

type fakeEvents struct{}

func (fakeEvents) ResponseCreated(responseID, vacancyID, ownerUserID int64, ownerEmail string, workerUserID int64, vacancyTitle string, shiftStartAt string, payAmount int64) (store.OutboxEventInput, error) {
	return store.OutboxEventInput{EventID: "event", Topic: "matching.response_created", EventType: "test", EventKey: ownerEmail, Payload: []byte("payload")}, nil
}

type fakeRepo struct {
	vacancy       store.Vacancy
	response      store.Response
	assignment    store.Assignment
	review        store.Review
	reply         store.ReviewReply
	replies       []store.ReviewReply
	rating        store.UserRating
	finalizeAdmin *store.AdminEventInput
}

func (r *fakeRepo) CreateVacancy(ctx context.Context, input store.Vacancy) (store.Vacancy, error) {
	input.ID = 1
	return input, nil
}

func (r *fakeRepo) GetVacancy(ctx context.Context, id int64) (store.Vacancy, error) {
	if r.vacancy.ID == 0 {
		return store.Vacancy{}, store.ErrNotFound
	}
	return r.vacancy, nil
}

func (r *fakeRepo) ListVacancies(ctx context.Context, filter store.ListVacancyFilter) ([]store.Vacancy, error) {
	return nil, nil
}

func (r *fakeRepo) CreateResponseWithOutbox(ctx context.Context, response store.Response, buildOutbox func(store.Response) (store.OutboxEventInput, error)) (store.Response, error) {
	response.ID = 1
	_, err := buildOutbox(response)
	return response, err
}

func (r *fakeRepo) GetResponse(ctx context.Context, id int64) (store.Response, error) {
	if r.response.ID == 0 {
		return store.Response{}, store.ErrNotFound
	}
	return r.response, nil
}

func (r *fakeRepo) ListResponsesForVacancy(ctx context.Context, vacancyID int64, limit uint64, offset uint64) ([]store.Response, error) {
	return nil, nil
}

func (r *fakeRepo) WithdrawResponse(ctx context.Context, responseID int64) (store.Response, error) {
	return r.response, nil
}

func (r *fakeRepo) AssignmentForResponse(ctx context.Context, responseID int64) (store.Assignment, error) {
	return store.Assignment{}, store.ErrNotFound
}

func (r *fakeRepo) AcceptResponse(ctx context.Context, input store.AssignmentSnapshotInput, admin *store.AdminEventInput) (store.Assignment, error) {
	return r.assignment, nil
}

func (r *fakeRepo) GetAssignment(ctx context.Context, id int64) (store.Assignment, error) {
	if r.assignment.ID == 0 {
		return store.Assignment{}, store.ErrNotFound
	}
	return r.assignment, nil
}

func (r *fakeRepo) FinalizeAssignment(ctx context.Context, id int64, status commonv1.AssignmentStatus, cancelledBy *int64, reason *string, admin *store.AdminEventInput) (store.Assignment, error) {
	r.finalizeAdmin = admin
	r.assignment.Status = status
	return r.assignment, nil
}

func (r *fakeRepo) CreateReviewAndRecalculateRating(ctx context.Context, review store.Review) (store.Review, error) {
	review.ID = 1
	return review, nil
}

func (r *fakeRepo) UpdateReviewAndRecalculateRating(ctx context.Context, id int64, rating *int32, body *string) (store.Review, error) {
	return r.review, nil
}

func (r *fakeRepo) GetReview(ctx context.Context, id int64) (store.Review, error) {
	if r.review.ID == 0 {
		return store.Review{}, store.ErrNotFound
	}
	return r.review, nil
}

func (r *fakeRepo) ListReviews(ctx context.Context, targetUserID *int64, limit uint64, offset uint64) ([]store.Review, error) {
	return nil, nil
}

func (r *fakeRepo) ListReviewReplies(ctx context.Context, reviewID int64, limit uint64, offset uint64) ([]store.ReviewReply, error) {
	return r.replies, nil
}

func (r *fakeRepo) GetUserRating(ctx context.Context, userID int64) (store.UserRating, error) {
	if r.rating.UserID == 0 {
		return store.UserRating{}, store.ErrNotFound
	}
	return r.rating, nil
}

func (r *fakeRepo) CreateReviewReply(ctx context.Context, reply store.ReviewReply) (store.ReviewReply, error) {
	reply.ID = 1
	return reply, nil
}

func (r *fakeRepo) GetReviewReply(ctx context.Context, id int64) (store.ReviewReply, error) {
	if r.reply.ID == 0 {
		return store.ReviewReply{}, store.ErrNotFound
	}
	return r.reply, nil
}

func (r *fakeRepo) UpdateReviewReply(ctx context.Context, id int64, body string) (store.ReviewReply, error) {
	r.reply.Body = body
	return r.reply, nil
}

func (r *fakeRepo) HasActiveVacancyForPvz(ctx context.Context, pvzID int64) (bool, error) {
	return false, nil
}

func (r *fakeRepo) CreateAdminEvent(ctx context.Context, input store.AdminEventInput) error {
	return nil
}
