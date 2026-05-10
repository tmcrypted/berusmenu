package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/tmcrypted/berusmenu/matching-service/internal/dto"
	"github.com/tmcrypted/berusmenu/matching-service/internal/events"
	"github.com/tmcrypted/berusmenu/matching-service/internal/store"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
)

const (
	defaultPageSize = uint64(20)
	maxPageSize     = uint64(100)
)

type Repository interface {
	CreateVacancy(ctx context.Context, input store.Vacancy) (store.Vacancy, error)
	GetVacancy(ctx context.Context, id int64) (store.Vacancy, error)
	ListVacancies(ctx context.Context, filter store.ListVacancyFilter) ([]store.Vacancy, error)
	CreateResponseWithOutbox(ctx context.Context, response store.Response, buildOutbox func(store.Response) (store.OutboxEventInput, error)) (store.Response, error)
	GetResponse(ctx context.Context, id int64) (store.Response, error)
	ListResponsesForVacancy(ctx context.Context, vacancyID int64, limit uint64, offset uint64) ([]store.Response, error)
	ListResponsesByWorker(ctx context.Context, workerID int64, limit uint64, offset uint64) ([]store.Response, error)
	WithdrawResponse(ctx context.Context, responseID int64) (store.Response, error)
	AssignmentForResponse(ctx context.Context, responseID int64) (store.Assignment, error)
	AcceptResponse(ctx context.Context, input store.AssignmentSnapshotInput, admin *store.AdminEventInput) (store.Assignment, error)
	GetAssignment(ctx context.Context, id int64) (store.Assignment, error)
	ListAssignmentsForUser(ctx context.Context, userID int64, limit uint64, offset uint64) ([]store.Assignment, error)
	AssignmentForVacancy(ctx context.Context, vacancyID int64) (store.Assignment, error)
	FinalizeAssignment(ctx context.Context, id int64, status commonv1.AssignmentStatus, cancelledBy *int64, reason *string, admin *store.AdminEventInput) (store.Assignment, error)
	CreateReviewAndRecalculateRating(ctx context.Context, review store.Review) (store.Review, error)
	UpdateReviewAndRecalculateRating(ctx context.Context, id int64, rating *int32, body *string) (store.Review, error)
	GetReview(ctx context.Context, id int64) (store.Review, error)
	ListReviews(ctx context.Context, targetUserID *int64, limit uint64, offset uint64) ([]store.Review, error)
	ListReviewReplies(ctx context.Context, reviewID int64, limit uint64, offset uint64) ([]store.ReviewReply, error)
	GetUserRating(ctx context.Context, userID int64) (store.UserRating, error)
	CreateReviewReply(ctx context.Context, reply store.ReviewReply) (store.ReviewReply, error)
	GetReviewReply(ctx context.Context, id int64) (store.ReviewReply, error)
	UpdateReviewReply(ctx context.Context, id int64, body string) (store.ReviewReply, error)
	HasActiveVacancyForPvz(ctx context.Context, pvzID int64) (bool, error)
	CreateAdminEvent(ctx context.Context, input store.AdminEventInput) error
}

type Inventory interface {
	CheckPvzForVacancy(ctx context.Context, pvzID int64) (dto.PvzSnapshot, error)
}

type Auth interface {
	GetUserByID(ctx context.Context, userID int64) (dto.UserSnapshot, error)
}

type EventBuilder interface {
	ResponseCreated(responseID, vacancyID, ownerUserID int64, ownerEmail string, workerUserID int64, vacancyTitle string, shiftStartAt string, payAmount int64) (store.OutboxEventInput, error)
}

type Service struct {
	repo      Repository
	inventory Inventory
	auth      Auth
	events    EventBuilder
	now       func() time.Time
}

func New(repo Repository, inventory Inventory, auth Auth, eventBuilder EventBuilder) *Service {
	return &Service{repo: repo, inventory: inventory, auth: auth, events: eventBuilder, now: time.Now}
}

func (s *Service) CreateVacancy(ctx context.Context, input dto.CreateVacancyInput) (store.Vacancy, error) {
	if err := requireRole(input.Actor, commonv1.UserRole_USER_ROLE_OWNER); err != nil {
		return store.Vacancy{}, err
	}
	if input.PvzID <= 0 {
		return store.Vacancy{}, serviceError(ErrInvalidArgument, "pvz_id is required")
	}
	if input.ShiftStartAt.IsZero() || input.ShiftEndAt.IsZero() {
		return store.Vacancy{}, serviceError(ErrInvalidArgument, "shift_start_at and shift_end_at are required")
	}
	if input.ShiftStartAt.Before(s.now().Add(-time.Minute)) {
		return store.Vacancy{}, serviceError(ErrInvalidArgument, "shift_start_at must be now or in the future")
	}
	if !input.ShiftEndAt.After(input.ShiftStartAt) {
		return store.Vacancy{}, serviceError(ErrInvalidArgument, "shift_end_at must be after shift_start_at")
	}
	if input.PayAmount <= 0 {
		return store.Vacancy{}, serviceError(ErrInvalidArgument, "pay_amount must be positive")
	}
	if input.SlotsTotal != 1 {
		return store.Vacancy{}, serviceError(ErrInvalidArgument, "slots_total must be 1 in MVP")
	}
	pvz, err := s.inventory.CheckPvzForVacancy(ctx, input.PvzID)
	if err != nil {
		return store.Vacancy{}, serviceError(ErrFailedPrecondition, "pvz is not available for vacancy")
	}
	if pvz.OwnerID != input.Actor.UserID {
		return store.Vacancy{}, serviceError(ErrPermissionDenied, "cannot create vacancy for another owner's pvz")
	}
	created, err := s.repo.CreateVacancy(ctx, store.Vacancy{
		OwnerID:            input.Actor.UserID,
		PvzID:              input.PvzID,
		Title:              trimOptional(input.Title),
		Description:        trimOptional(input.Description),
		ShiftStartAt:       input.ShiftStartAt,
		ShiftEndAt:         input.ShiftEndAt,
		PvzNameSnapshot:    pvz.CompanyName,
		PvzAddressSnapshot: pvz.Address,
		PayAmount:          input.PayAmount,
		SlotsTotal:         input.SlotsTotal,
	})
	if err != nil {
		return store.Vacancy{}, serviceError(ErrInternal, "failed to create vacancy")
	}
	return created, nil
}

func (s *Service) GetVacancy(ctx context.Context, input dto.IDInput) (store.Vacancy, error) {
	vacancy, err := s.repo.GetVacancy(ctx, input.ID)
	if err != nil {
		return store.Vacancy{}, notFoundOrInternal(err, "vacancy not found", "failed to get vacancy")
	}
	if err := canReadVacancy(input.Actor, vacancy); err != nil {
		return store.Vacancy{}, err
	}
	return vacancy, nil
}

func (s *Service) ListVacancies(ctx context.Context, input dto.ListVacanciesInput) ([]store.Vacancy, *string, error) {
	if input.Actor.UserID <= 0 {
		return nil, nil, serviceError(ErrUnauthenticated, "missing authenticated user")
	}
	page, err := pageFrom(input.PageSize, input.PageToken)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.repo.ListVacancies(ctx, store.ListVacancyFilter{
		ActorUserID:    input.Actor.UserID,
		Role:           input.Actor.Role,
		PvzID:          input.PvzID,
		ShiftStartFrom: input.ShiftStartFrom,
		ShiftStartTo:   input.ShiftStartTo,
		Status:         input.Status,
		Limit:          page.Limit + 1,
		Offset:         page.Offset,
	})
	if err != nil {
		return nil, nil, serviceError(ErrInternal, "failed to list vacancies")
	}
	return trimPage(items, page)
}

func (s *Service) CreateResponse(ctx context.Context, input dto.CreateResponseInput) (store.Response, error) {
	if err := requireRole(input.Actor, commonv1.UserRole_USER_ROLE_WORKER); err != nil {
		return store.Response{}, err
	}
	vacancy, err := s.repo.GetVacancy(ctx, input.VacancyID)
	if err != nil {
		return store.Response{}, notFoundOrInternal(err, "vacancy not found", "failed to get vacancy")
	}
	if vacancy.Status != commonv1.VacancyStatus_VACANCY_STATUS_PUBLISHED {
		return store.Response{}, serviceError(ErrFailedPrecondition, "vacancy is not published")
	}
	if vacancy.OwnerID == input.Actor.UserID {
		return store.Response{}, serviceError(ErrFailedPrecondition, "worker cannot respond to own vacancy")
	}
	owner, err := s.auth.GetUserByID(ctx, vacancy.OwnerID)
	if err != nil {
		return store.Response{}, serviceError(ErrUnavailable, "failed to load owner snapshot")
	}
	title := deref(vacancy.Title)
	if title == "" {
		title = "PVZ shift"
	}
	created, err := s.repo.CreateResponseWithOutbox(ctx, store.Response{VacancyID: vacancy.ID, WorkerID: input.Actor.UserID, Message: trimOptional(input.Message)}, func(created store.Response) (store.OutboxEventInput, error) {
		return s.events.ResponseCreated(created.ID, vacancy.ID, vacancy.OwnerID, owner.Email, input.Actor.UserID, title, vacancy.ShiftStartAt.Format(time.RFC3339), vacancy.PayAmount)
	})
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return store.Response{}, serviceError(ErrAlreadyExists, "response already exists for vacancy")
		}
		return store.Response{}, serviceError(ErrInternal, "failed to create response")
	}
	return created, nil
}

func (s *Service) WithdrawResponse(ctx context.Context, input dto.WithdrawResponseInput) (store.Response, error) {
	response, err := s.repo.GetResponse(ctx, input.ResponseID)
	if err != nil {
		return store.Response{}, notFoundOrInternal(err, "response not found", "failed to get response")
	}
	if response.WorkerID != input.Actor.UserID {
		return store.Response{}, serviceError(ErrPermissionDenied, "cannot withdraw another worker's response")
	}
	if _, err := s.repo.AssignmentForResponse(ctx, response.ID); err == nil {
		return store.Response{}, serviceError(ErrFailedPrecondition, "cannot withdraw response after assignment")
	} else if !errors.Is(err, store.ErrNotFound) {
		return store.Response{}, serviceError(ErrInternal, "failed to check assignment")
	}
	updated, err := s.repo.WithdrawResponse(ctx, response.ID)
	if err != nil {
		return store.Response{}, notFoundOrInternal(err, "pending response not found", "failed to withdraw response")
	}
	return updated, nil
}

func (s *Service) ListResponsesForVacancy(ctx context.Context, input dto.ListResponsesInput) ([]store.Response, *string, error) {
	vacancy, err := s.repo.GetVacancy(ctx, input.VacancyID)
	if err != nil {
		return nil, nil, notFoundOrInternal(err, "vacancy not found", "failed to get vacancy")
	}
	if input.Actor.Role != commonv1.UserRole_USER_ROLE_MODERATOR && vacancy.OwnerID != input.Actor.UserID {
		return nil, nil, serviceError(ErrPermissionDenied, "cannot list responses for vacancy")
	}
	page, err := pageFrom(input.PageSize, input.PageToken)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.repo.ListResponsesForVacancy(ctx, vacancy.ID, page.Limit+1, page.Offset)
	if err != nil {
		return nil, nil, serviceError(ErrInternal, "failed to list responses")
	}
	return trimPage(items, page)
}

func (s *Service) ListMyResponses(ctx context.Context, input dto.ListMyResponsesInput) ([]store.Response, *string, error) {
	if err := requireRole(input.Actor, commonv1.UserRole_USER_ROLE_WORKER); err != nil {
		return nil, nil, err
	}
	page, err := pageFrom(input.PageSize, input.PageToken)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.repo.ListResponsesByWorker(ctx, input.Actor.UserID, page.Limit+1, page.Offset)
	if err != nil {
		return nil, nil, serviceError(ErrInternal, "failed to list responses")
	}
	return trimPage(items, page)
}

func (s *Service) AcceptResponse(ctx context.Context, input dto.AcceptResponseInput) (store.Assignment, error) {
	response, err := s.repo.GetResponse(ctx, input.ResponseID)
	if err != nil {
		return store.Assignment{}, notFoundOrInternal(err, "response not found", "failed to get response")
	}
	vacancy, err := s.repo.GetVacancy(ctx, response.VacancyID)
	if err != nil {
		return store.Assignment{}, notFoundOrInternal(err, "vacancy not found", "failed to get vacancy")
	}
	if input.Actor.Role != commonv1.UserRole_USER_ROLE_MODERATOR && vacancy.OwnerID != input.Actor.UserID {
		return store.Assignment{}, serviceError(ErrPermissionDenied, "cannot accept response")
	}
	if response.Status != commonv1.ResponseStatus_RESPONSE_STATUS_PENDING || vacancy.Status != commonv1.VacancyStatus_VACANCY_STATUS_PUBLISHED {
		return store.Assignment{}, serviceError(ErrFailedPrecondition, "response cannot be accepted")
	}
	owner, err := s.auth.GetUserByID(ctx, vacancy.OwnerID)
	if err != nil {
		return store.Assignment{}, serviceError(ErrUnavailable, "failed to load owner snapshot")
	}
	worker, err := s.auth.GetUserByID(ctx, response.WorkerID)
	if err != nil {
		return store.Assignment{}, serviceError(ErrUnavailable, "failed to load worker snapshot")
	}
	var admin *store.AdminEventInput
	if input.Actor.Role == commonv1.UserRole_USER_ROLE_MODERATOR {
		reason, err := required(deref(input.AdminReason), "reason")
		if err != nil {
			return store.Assignment{}, err
		}
		previous := "pending"
		next := "accepted"
		admin = &store.AdminEventInput{
			ActorUserID:   input.Actor.UserID,
			TargetType:    "response",
			TargetID:      response.ID,
			Action:        "accept_response",
			PreviousState: &previous,
			NewState:      &next,
			Reason:        reason,
		}
	}
	assignment, err := s.repo.AcceptResponse(ctx, store.AssignmentSnapshotInput{
		ResponseID:         response.ID,
		Vacancy:            vacancy,
		WorkerID:           response.WorkerID,
		PvzNameSnapshot:    vacancy.PvzNameSnapshot,
		PvzAddressSnapshot: vacancy.PvzAddressSnapshot,
		OwnerNameSnapshot:  owner.FullName,
		WorkerNameSnapshot: worker.FullName,
	}, admin)
	if err != nil {
		return store.Assignment{}, notFoundOrInternal(err, "response cannot be accepted", "failed to accept response")
	}
	return assignment, nil
}

func (s *Service) GetAssignment(ctx context.Context, input dto.IDInput) (store.Assignment, error) {
	assignment, err := s.repo.GetAssignment(ctx, input.ID)
	if err != nil {
		return store.Assignment{}, notFoundOrInternal(err, "assignment not found", "failed to get assignment")
	}
	if err := canReadAssignment(input.Actor, assignment); err != nil {
		return store.Assignment{}, err
	}
	return assignment, nil
}

func (s *Service) ListMyAssignments(ctx context.Context, input dto.ListMyAssignmentsInput) ([]store.Assignment, *string, error) {
	if input.Actor.UserID <= 0 {
		return nil, nil, serviceError(ErrUnauthenticated, "missing authenticated user")
	}
	if input.Actor.Role == commonv1.UserRole_USER_ROLE_MODERATOR {
		return nil, nil, serviceError(ErrPermissionDenied, "moderator does not have personal assignments")
	}
	page, err := pageFrom(input.PageSize, input.PageToken)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.repo.ListAssignmentsForUser(ctx, input.Actor.UserID, page.Limit+1, page.Offset)
	if err != nil {
		return nil, nil, serviceError(ErrInternal, "failed to list assignments")
	}
	return trimPage(items, page)
}

func (s *Service) GetAssignmentForVacancy(ctx context.Context, input dto.GetAssignmentForVacancyInput) (store.Assignment, error) {
	vacancy, err := s.repo.GetVacancy(ctx, input.VacancyID)
	if err != nil {
		return store.Assignment{}, notFoundOrInternal(err, "vacancy not found", "failed to get vacancy")
	}
	if input.Actor.Role != commonv1.UserRole_USER_ROLE_MODERATOR && vacancy.OwnerID != input.Actor.UserID {
		return store.Assignment{}, serviceError(ErrPermissionDenied, "cannot access assignment for vacancy")
	}
	assignment, err := s.repo.AssignmentForVacancy(ctx, vacancy.ID)
	if err != nil {
		return store.Assignment{}, notFoundOrInternal(err, "assignment not found", "failed to get assignment")
	}
	return assignment, nil
}

func (s *Service) CancelAssignment(ctx context.Context, input dto.CancelAssignmentInput) (store.Assignment, error) {
	assignment, err := s.repo.GetAssignment(ctx, input.AssignmentID)
	if err != nil {
		return store.Assignment{}, notFoundOrInternal(err, "assignment not found", "failed to get assignment")
	}
	if input.Actor.Role != commonv1.UserRole_USER_ROLE_MODERATOR && input.Actor.UserID != assignment.OwnerID && input.Actor.UserID != assignment.WorkerID {
		return store.Assignment{}, serviceError(ErrPermissionDenied, "cannot cancel assignment")
	}
	cancelledBy := input.Actor.UserID
	var admin *store.AdminEventInput
	if input.Actor.Role == commonv1.UserRole_USER_ROLE_MODERATOR {
		reason, err := required(deref(input.AdminReason), "reason")
		if err != nil {
			return store.Assignment{}, err
		}
		previous := "active"
		next := "cancelled"
		admin = &store.AdminEventInput{
			ActorUserID:   input.Actor.UserID,
			TargetType:    "assignment",
			TargetID:      assignment.ID,
			Action:        "cancel_assignment",
			PreviousState: &previous,
			NewState:      &next,
			Reason:        reason,
		}
	}
	updated, err := s.repo.FinalizeAssignment(ctx, assignment.ID, commonv1.AssignmentStatus_ASSIGNMENT_STATUS_CANCELLED, &cancelledBy, trimOptional(input.Reason), admin)
	if err != nil {
		return store.Assignment{}, notFoundOrInternal(err, "active assignment not found", "failed to cancel assignment")
	}
	return updated, nil
}

func (s *Service) CompleteAssignment(ctx context.Context, input dto.CompleteAssignmentInput) (store.Assignment, error) {
	assignment, err := s.repo.GetAssignment(ctx, input.AssignmentID)
	if err != nil {
		return store.Assignment{}, notFoundOrInternal(err, "assignment not found", "failed to get assignment")
	}
	if input.Actor.Role != commonv1.UserRole_USER_ROLE_MODERATOR && input.Actor.UserID != assignment.OwnerID {
		return store.Assignment{}, serviceError(ErrPermissionDenied, "only owner can complete assignment")
	}
	var admin *store.AdminEventInput
	if input.Actor.Role == commonv1.UserRole_USER_ROLE_MODERATOR {
		reason, err := required(deref(input.AdminReason), "reason")
		if err != nil {
			return store.Assignment{}, err
		}
		previous := "active"
		next := "completed"
		admin = &store.AdminEventInput{
			ActorUserID:   input.Actor.UserID,
			TargetType:    "assignment",
			TargetID:      assignment.ID,
			Action:        "complete_assignment",
			PreviousState: &previous,
			NewState:      &next,
			Reason:        reason,
		}
	}
	updated, err := s.repo.FinalizeAssignment(ctx, assignment.ID, commonv1.AssignmentStatus_ASSIGNMENT_STATUS_COMPLETED, nil, nil, admin)
	if err != nil {
		return store.Assignment{}, notFoundOrInternal(err, "active assignment not found", "failed to complete assignment")
	}
	return updated, nil
}

func (s *Service) CreateReview(ctx context.Context, input dto.CreateReviewInput) (store.Review, error) {
	assignment, err := s.repo.GetAssignment(ctx, input.AssignmentID)
	if err != nil {
		return store.Review{}, notFoundOrInternal(err, "assignment not found", "failed to get assignment")
	}
	if assignment.Status != commonv1.AssignmentStatus_ASSIGNMENT_STATUS_COMPLETED && assignment.Status != commonv1.AssignmentStatus_ASSIGNMENT_STATUS_CANCELLED {
		return store.Review{}, serviceError(ErrFailedPrecondition, "assignment is not finalized")
	}
	targetID, targetRole, err := reviewTarget(input.Actor, assignment)
	if err != nil {
		return store.Review{}, err
	}
	if input.Rating < 1 || input.Rating > 5 {
		return store.Review{}, serviceError(ErrInvalidArgument, "rating must be between 1 and 5")
	}
	created, err := s.repo.CreateReviewAndRecalculateRating(ctx, store.Review{
		AssignmentID: assignment.ID,
		AuthorUserID: input.Actor.UserID,
		TargetUserID: targetID,
		TargetRole:   targetRole,
		Rating:       input.Rating,
		Body:         trimOptional(input.Body),
	})
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return store.Review{}, serviceError(ErrAlreadyExists, "review already exists for assignment")
		}
		return store.Review{}, serviceError(ErrInternal, "failed to create review")
	}
	return created, nil
}

func (s *Service) UpdateReview(ctx context.Context, input dto.UpdateReviewInput) (store.Review, error) {
	current, err := s.repo.GetReview(ctx, input.ReviewID)
	if err != nil {
		return store.Review{}, notFoundOrInternal(err, "review not found", "failed to get review")
	}
	if current.AuthorUserID != input.Actor.UserID && input.Actor.Role != commonv1.UserRole_USER_ROLE_MODERATOR {
		return store.Review{}, serviceError(ErrPermissionDenied, "cannot update review")
	}
	if input.Rating != nil && (*input.Rating < 1 || *input.Rating > 5) {
		return store.Review{}, serviceError(ErrInvalidArgument, "rating must be between 1 and 5")
	}
	if input.Rating == nil && input.Body == nil {
		return store.Review{}, serviceError(ErrInvalidArgument, "at least one field is required")
	}
	return s.repo.UpdateReviewAndRecalculateRating(ctx, current.ID, input.Rating, trimOptional(input.Body))
}

func (s *Service) CreateReviewReply(ctx context.Context, input dto.CreateReviewReplyInput) (store.ReviewReply, error) {
	body, err := required(input.Body, "body")
	if err != nil {
		return store.ReviewReply{}, err
	}
	review, err := s.repo.GetReview(ctx, input.ReviewID)
	if err != nil {
		return store.ReviewReply{}, notFoundOrInternal(err, "review not found", "failed to get review")
	}
	assignment, err := s.repo.GetAssignment(ctx, review.AssignmentID)
	if err != nil {
		return store.ReviewReply{}, notFoundOrInternal(err, "assignment not found", "failed to get assignment")
	}
	if input.Actor.UserID != assignment.OwnerID && input.Actor.UserID != assignment.WorkerID && input.Actor.Role != commonv1.UserRole_USER_ROLE_MODERATOR {
		return store.ReviewReply{}, serviceError(ErrPermissionDenied, "cannot reply to review")
	}
	if input.TargetUserID != assignment.OwnerID && input.TargetUserID != assignment.WorkerID {
		return store.ReviewReply{}, serviceError(ErrInvalidArgument, "target_user_id must be an assignment participant")
	}
	return s.repo.CreateReviewReply(ctx, store.ReviewReply{ReviewID: review.ID, AuthorUserID: input.Actor.UserID, TargetUserID: input.TargetUserID, Body: body})
}

func (s *Service) UpdateReviewReply(ctx context.Context, input dto.UpdateReviewReplyInput) (store.ReviewReply, error) {
	body, err := required(input.Body, "body")
	if err != nil {
		return store.ReviewReply{}, err
	}
	reply, err := s.repo.GetReviewReply(ctx, input.ReplyID)
	if err != nil {
		return store.ReviewReply{}, notFoundOrInternal(err, "reply not found", "failed to get reply")
	}
	if reply.AuthorUserID != input.Actor.UserID && input.Actor.Role != commonv1.UserRole_USER_ROLE_MODERATOR {
		return store.ReviewReply{}, serviceError(ErrPermissionDenied, "cannot update reply")
	}
	return s.repo.UpdateReviewReply(ctx, reply.ID, body)
}

func (s *Service) ListReviews(ctx context.Context, input dto.ListReviewsInput) ([]store.Review, *string, error) {
	page, err := pageFrom(input.PageSize, input.PageToken)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.repo.ListReviews(ctx, input.TargetUserID, page.Limit+1, page.Offset)
	if err != nil {
		return nil, nil, serviceError(ErrInternal, "failed to list reviews")
	}
	return trimPage(items, page)
}

func (s *Service) ListReviewReplies(ctx context.Context, input dto.ListReviewRepliesInput) ([]store.ReviewReply, *string, error) {
	review, err := s.repo.GetReview(ctx, input.ReviewID)
	if err != nil {
		return nil, nil, notFoundOrInternal(err, "review not found", "failed to get review")
	}
	assignment, err := s.repo.GetAssignment(ctx, review.AssignmentID)
	if err != nil {
		return nil, nil, notFoundOrInternal(err, "assignment not found", "failed to get assignment")
	}
	if input.Actor.Role != commonv1.UserRole_USER_ROLE_MODERATOR && input.Actor.UserID != assignment.OwnerID && input.Actor.UserID != assignment.WorkerID {
		return nil, nil, serviceError(ErrPermissionDenied, "cannot list review replies")
	}
	page, err := pageFrom(input.PageSize, input.PageToken)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.repo.ListReviewReplies(ctx, review.ID, page.Limit+1, page.Offset)
	if err != nil {
		return nil, nil, serviceError(ErrInternal, "failed to list review replies")
	}
	return trimPage(items, page)
}

func (s *Service) GetUserRating(ctx context.Context, input dto.GetUserRatingInput) (store.UserRating, error) {
	if input.UserID <= 0 {
		return store.UserRating{}, serviceError(ErrInvalidArgument, "user_id is required")
	}
	rating, err := s.repo.GetUserRating(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.UserRating{UserID: input.UserID}, nil
		}
		return store.UserRating{}, serviceError(ErrInternal, "failed to get user rating")
	}
	return rating, nil
}

func (s *Service) HasActiveVacancyForPvz(ctx context.Context, pvzID int64) (bool, error) {
	if pvzID <= 0 {
		return false, serviceError(ErrInvalidArgument, "pvz_id is required")
	}
	return s.repo.HasActiveVacancyForPvz(ctx, pvzID)
}

func (s *Service) CheckResponseChatWritable(ctx context.Context, responseID int64) (bool, *store.Assignment, error) {
	if responseID <= 0 {
		return false, nil, serviceError(ErrInvalidArgument, "response_id is required")
	}
	assignment, err := s.repo.AssignmentForResponse(ctx, responseID)
	if errors.Is(err, store.ErrNotFound) {
		return true, nil, nil
	}
	if err != nil {
		return false, nil, serviceError(ErrInternal, "failed to check assignment")
	}
	writable := assignment.Status == commonv1.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE
	return writable, &assignment, nil
}

func requireRole(actor dto.Actor, role commonv1.UserRole) error {
	if actor.UserID <= 0 {
		return serviceError(ErrUnauthenticated, "missing authenticated user")
	}
	if actor.Role != role {
		return serviceError(ErrPermissionDenied, "insufficient role")
	}
	return nil
}

func canReadVacancy(actor dto.Actor, vacancy store.Vacancy) error {
	if actor.Role == commonv1.UserRole_USER_ROLE_MODERATOR {
		return nil
	}
	if actor.Role == commonv1.UserRole_USER_ROLE_OWNER && actor.UserID == vacancy.OwnerID {
		return nil
	}
	if actor.Role == commonv1.UserRole_USER_ROLE_WORKER && vacancy.Status == commonv1.VacancyStatus_VACANCY_STATUS_PUBLISHED {
		return nil
	}
	return serviceError(ErrPermissionDenied, "cannot access vacancy")
}

func canReadAssignment(actor dto.Actor, assignment store.Assignment) error {
	if actor.Role == commonv1.UserRole_USER_ROLE_MODERATOR {
		return nil
	}
	if actor.UserID == assignment.OwnerID || actor.UserID == assignment.WorkerID {
		return nil
	}
	return serviceError(ErrPermissionDenied, "cannot access assignment")
}

func reviewTarget(actor dto.Actor, assignment store.Assignment) (int64, commonv1.ReviewTargetRole, error) {
	if actor.UserID == assignment.OwnerID {
		return assignment.WorkerID, commonv1.ReviewTargetRole_REVIEW_TARGET_ROLE_WORKER, nil
	}
	if actor.UserID == assignment.WorkerID {
		return assignment.OwnerID, commonv1.ReviewTargetRole_REVIEW_TARGET_ROLE_OWNER, nil
	}
	return 0, commonv1.ReviewTargetRole_REVIEW_TARGET_ROLE_UNSPECIFIED, serviceError(ErrPermissionDenied, "only assignment participants can review")
}

func required(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", serviceError(ErrInvalidArgument, field+" is required")
	}
	return trimmed, nil
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func pageFrom(pageSize int32, token *string) (dto.Page, error) {
	limit := defaultPageSize
	if pageSize > 0 {
		limit = uint64(pageSize)
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	var offset uint64
	if token != nil && strings.TrimSpace(*token) != "" {
		parsed, err := strconv.ParseUint(strings.TrimSpace(*token), 10, 64)
		if err != nil {
			return dto.Page{}, serviceError(ErrInvalidArgument, "invalid page token")
		}
		offset = parsed
	}
	return dto.Page{Limit: limit, Offset: offset}, nil
}

func trimPage[T any](items []T, page dto.Page) ([]T, *string, error) {
	if uint64(len(items)) <= page.Limit {
		return items, nil, nil
	}
	trimmed := items[:page.Limit]
	next := strconv.FormatUint(page.Offset+page.Limit, 10)
	return trimmed, &next, nil
}

func notFoundOrInternal(err error, notFoundMessage string, internalMessage string) error {
	if errors.Is(err, store.ErrNotFound) {
		return serviceError(ErrNotFound, notFoundMessage)
	}
	return serviceError(ErrInternal, internalMessage)
}

var _ EventBuilder = (*events.Builder)(nil)
