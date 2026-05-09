package server

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/tmcrypted/berusmenu/matching-service/internal/dto"
	"github.com/tmcrypted/berusmenu/matching-service/internal/service"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	matchingv1 "github.com/tmcrypted/berusmenu/proto/gen/go/matching/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Dependencies struct {
	Repo      service.Repository
	Inventory service.Inventory
	Auth      service.Auth
	Events    service.EventBuilder
}

type Server struct {
	matchingv1.UnimplementedMatchingServiceServer
	matching *service.Service
}

func New(deps Dependencies) *Server {
	return &Server{matching: service.New(deps.Repo, deps.Inventory, deps.Auth, deps.Events)}
}

func (s *Server) CreateVacancy(ctx context.Context, req *matchingv1.CreateVacancyRequest) (*matchingv1.CreateVacancyResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	vacancy, err := s.matching.CreateVacancy(ctx, dto.CreateVacancyInput{
		Actor:        actor,
		PvzID:        req.GetPvzId(),
		Title:        req.Title,
		Description:  req.Description,
		ShiftStartAt: req.GetShiftStartAt().AsTime(),
		ShiftEndAt:   req.GetShiftEndAt().AsTime(),
		PayAmount:    req.GetPayAmount(),
		SlotsTotal:   req.GetSlotsTotal(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.CreateVacancyResponse{Vacancy: toProtoVacancy(vacancy)}, nil
}

func (s *Server) GetVacancy(ctx context.Context, req *matchingv1.GetVacancyRequest) (*matchingv1.GetVacancyResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	vacancy, err := s.matching.GetVacancy(ctx, dto.IDInput{Actor: actor, ID: req.GetId()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.GetVacancyResponse{Vacancy: toProtoVacancy(vacancy)}, nil
}

func (s *Server) ListVacancies(ctx context.Context, req *matchingv1.ListVacanciesRequest) (*matchingv1.ListVacanciesResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	var from, to *timeWrapper
	if req.ShiftStartFrom != nil {
		from = &timeWrapper{req.ShiftStartFrom.AsTime()}
	}
	if req.ShiftStartTo != nil {
		to = &timeWrapper{req.ShiftStartTo.AsTime()}
	}
	items, next, err := s.matching.ListVacancies(ctx, dto.ListVacanciesInput{
		Actor:          actor,
		PvzID:          req.PvzId,
		ShiftStartFrom: unwrapTime(from),
		ShiftStartTo:   unwrapTime(to),
		Status:         req.GetStatus(),
		PageSize:       req.GetPageSize(),
		PageToken:      req.PageToken,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.ListVacanciesResponse{Items: toProtoVacancies(items), NextPageToken: next}, nil
}

func (s *Server) ListOwnerVacancies(ctx context.Context, req *matchingv1.ListOwnerVacanciesRequest) (*matchingv1.ListOwnerVacanciesResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	items, next, err := s.matching.ListVacancies(ctx, dto.ListVacanciesInput{Actor: actor, Status: req.GetStatus(), PageSize: req.GetPageSize(), PageToken: req.PageToken})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.ListOwnerVacanciesResponse{Items: toProtoVacancies(items), NextPageToken: next}, nil
}

func (s *Server) CreateResponse(ctx context.Context, req *matchingv1.CreateResponseRequest) (*matchingv1.CreateResponseResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	response, err := s.matching.CreateResponse(ctx, dto.CreateResponseInput{Actor: actor, VacancyID: req.GetVacancyId(), Message: req.Message})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.CreateResponseResponse{Response: toProtoResponse(response)}, nil
}

func (s *Server) WithdrawResponse(ctx context.Context, req *matchingv1.WithdrawResponseRequest) (*matchingv1.WithdrawResponseResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	response, err := s.matching.WithdrawResponse(ctx, dto.WithdrawResponseInput{Actor: actor, ResponseID: req.GetResponseId()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.WithdrawResponseResponse{Response: toProtoResponse(response)}, nil
}

func (s *Server) ListResponsesForVacancy(ctx context.Context, req *matchingv1.ListResponsesForVacancyRequest) (*matchingv1.ListResponsesForVacancyResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	items, next, err := s.matching.ListResponsesForVacancy(ctx, dto.ListResponsesInput{Actor: actor, VacancyID: req.GetVacancyId(), PageSize: req.GetPageSize(), PageToken: req.PageToken})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.ListResponsesForVacancyResponse{Items: toProtoResponses(items), NextPageToken: next}, nil
}

func (s *Server) AcceptResponse(ctx context.Context, req *matchingv1.AcceptResponseRequest) (*matchingv1.AcceptResponseResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	assignment, err := s.matching.AcceptResponse(ctx, dto.AcceptResponseInput{
		Actor:       actor,
		ResponseID:  req.GetResponseId(),
		AdminReason: req.AdminReason,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.AcceptResponseResponse{Assignment: toProtoAssignment(assignment)}, nil
}

func (s *Server) CancelAssignment(ctx context.Context, req *matchingv1.CancelAssignmentRequest) (*matchingv1.CancelAssignmentResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	assignment, err := s.matching.CancelAssignment(ctx, dto.CancelAssignmentInput{
		Actor:        actor,
		AssignmentID: req.GetAssignmentId(),
		Reason:       req.CancelReason,
		AdminReason:  req.AdminReason,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.CancelAssignmentResponse{Assignment: toProtoAssignment(assignment)}, nil
}

func (s *Server) CompleteAssignment(ctx context.Context, req *matchingv1.CompleteAssignmentRequest) (*matchingv1.CompleteAssignmentResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	assignment, err := s.matching.CompleteAssignment(ctx, dto.CompleteAssignmentInput{
		Actor:        actor,
		AssignmentID: req.GetAssignmentId(),
		AdminReason:  req.AdminReason,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.CompleteAssignmentResponse{Assignment: toProtoAssignment(assignment)}, nil
}

func (s *Server) CreateReview(ctx context.Context, req *matchingv1.CreateReviewRequest) (*matchingv1.CreateReviewResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	review, err := s.matching.CreateReview(ctx, dto.CreateReviewInput{Actor: actor, AssignmentID: req.GetAssignmentId(), Rating: req.GetRating(), Body: req.Body})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.CreateReviewResponse{Review: toProtoReview(review)}, nil
}

func (s *Server) UpdateReview(ctx context.Context, req *matchingv1.UpdateReviewRequest) (*matchingv1.UpdateReviewResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	review, err := s.matching.UpdateReview(ctx, dto.UpdateReviewInput{Actor: actor, ReviewID: req.GetReviewId(), Rating: req.Rating, Body: req.Body})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.UpdateReviewResponse{Review: toProtoReview(review)}, nil
}

func (s *Server) CreateReviewReply(ctx context.Context, req *matchingv1.CreateReviewReplyRequest) (*matchingv1.CreateReviewReplyResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	reply, err := s.matching.CreateReviewReply(ctx, dto.CreateReviewReplyInput{Actor: actor, ReviewID: req.GetReviewId(), TargetUserID: req.GetTargetUserId(), Body: req.GetBody()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.CreateReviewReplyResponse{Reply: toProtoReviewReply(reply)}, nil
}

func (s *Server) UpdateReviewReply(ctx context.Context, req *matchingv1.UpdateReviewReplyRequest) (*matchingv1.UpdateReviewReplyResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	reply, err := s.matching.UpdateReviewReply(ctx, dto.UpdateReviewReplyInput{Actor: actor, ReplyID: req.GetReplyId(), Body: req.GetBody()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.UpdateReviewReplyResponse{Reply: toProtoReviewReply(reply)}, nil
}

func (s *Server) ListReviewReplies(ctx context.Context, req *matchingv1.ListReviewRepliesRequest) (*matchingv1.ListReviewRepliesResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	items, next, err := s.matching.ListReviewReplies(ctx, dto.ListReviewRepliesInput{
		Actor:     actor,
		ReviewID:  req.GetReviewId(),
		PageSize:  req.GetPageSize(),
		PageToken: req.PageToken,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.ListReviewRepliesResponse{Items: toProtoReviewReplies(items), NextPageToken: next}, nil
}

func (s *Server) ListReviews(ctx context.Context, req *matchingv1.ListReviewsRequest) (*matchingv1.ListReviewsResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	items, next, err := s.matching.ListReviews(ctx, dto.ListReviewsInput{Actor: actor, TargetUserID: req.TargetUserId, PageSize: req.GetPageSize(), PageToken: req.PageToken})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.ListReviewsResponse{Items: toProtoReviews(items), NextPageToken: next}, nil
}

func (s *Server) GetUserRating(ctx context.Context, req *matchingv1.GetUserRatingRequest) (*matchingv1.GetUserRatingResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	rating, err := s.matching.GetUserRating(ctx, dto.GetUserRatingInput{Actor: actor, UserID: req.GetUserId()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.GetUserRatingResponse{Rating: toProtoUserRating(rating)}, nil
}

func (s *Server) HasActiveVacancyForPvz(ctx context.Context, req *matchingv1.HasActiveVacancyForPvzRequest) (*matchingv1.HasActiveVacancyForPvzResponse, error) {
	hasActive, err := s.matching.HasActiveVacancyForPvz(ctx, req.GetPvzId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &matchingv1.HasActiveVacancyForPvzResponse{HasActiveVacancy: hasActive}, nil
}

func (s *Server) CheckResponseChatWritable(ctx context.Context, req *matchingv1.CheckResponseChatWritableRequest) (*matchingv1.CheckResponseChatWritableResponse, error) {
	writable, assignment, err := s.matching.CheckResponseChatWritable(ctx, req.GetResponseId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	res := &matchingv1.CheckResponseChatWritableResponse{Writable: writable}
	if assignment != nil {
		res.AssignmentId = &assignment.ID
		res.AssignmentStatus = &assignment.Status
	}
	return res, nil
}

type timeWrapper struct{ value time.Time }

func unwrapTime(value *timeWrapper) *time.Time {
	if value == nil {
		return nil
	}
	return &value.value
}

func actorFromMetadata(ctx context.Context) (dto.Actor, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return dto.Actor{}, status.Error(codes.Unauthenticated, "missing identity metadata")
	}
	userID, err := metadataInt64(md, "x-user-id")
	if err != nil {
		return dto.Actor{}, err
	}
	roles := md.Get("x-user-role")
	if len(roles) == 0 {
		return dto.Actor{}, status.Error(codes.Unauthenticated, "missing identity metadata")
	}
	role := roleFromString(roles[0])
	if role == commonv1.UserRole_USER_ROLE_UNSPECIFIED {
		return dto.Actor{}, status.Error(codes.Unauthenticated, "invalid identity metadata")
	}
	email := ""
	if values := md.Get("x-user-email"); len(values) > 0 {
		email = values[0]
	}
	return dto.Actor{UserID: userID, Role: role, Email: email}, nil
}

func metadataInt64(md metadata.MD, key string) (int64, error) {
	values := md.Get(key)
	if len(values) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing identity metadata")
	}
	value, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || value <= 0 {
		return 0, status.Error(codes.Unauthenticated, "invalid identity metadata")
	}
	return value, nil
}

func roleFromString(value string) commonv1.UserRole {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "owner":
		return commonv1.UserRole_USER_ROLE_OWNER
	case "worker":
		return commonv1.UserRole_USER_ROLE_WORKER
	case "moderator":
		return commonv1.UserRole_USER_ROLE_MODERATOR
	default:
		return commonv1.UserRole_USER_ROLE_UNSPECIFIED
	}
}

func toGRPCError(err error) error {
	var svcErr service.Error
	if !errors.As(err, &svcErr) {
		return status.Error(codes.Internal, "internal")
	}
	code := codes.Internal
	switch {
	case errors.Is(svcErr.Kind, service.ErrInvalidArgument):
		code = codes.InvalidArgument
	case errors.Is(svcErr.Kind, service.ErrAlreadyExists):
		code = codes.AlreadyExists
	case errors.Is(svcErr.Kind, service.ErrUnauthenticated):
		code = codes.Unauthenticated
	case errors.Is(svcErr.Kind, service.ErrPermissionDenied):
		code = codes.PermissionDenied
	case errors.Is(svcErr.Kind, service.ErrNotFound):
		code = codes.NotFound
	case errors.Is(svcErr.Kind, service.ErrFailedPrecondition):
		code = codes.FailedPrecondition
	case errors.Is(svcErr.Kind, service.ErrUnavailable):
		code = codes.Unavailable
	}
	return status.Error(code, svcErr.Message)
}
