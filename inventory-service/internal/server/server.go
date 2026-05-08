package server

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/tmcrypted/berusmenu/inventory-service/internal/dto"
	"github.com/tmcrypted/berusmenu/inventory-service/internal/service"
	"github.com/tmcrypted/berusmenu/inventory-service/internal/store"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	inventoryv1 "github.com/tmcrypted/berusmenu/proto/gen/go/inventory/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Dependencies struct {
	Pvz             service.PvzRepository
	ActiveVacancies service.ActiveVacancyChecker
}

type Server struct {
	inventoryv1.UnimplementedInventoryServiceServer
	inventory *service.Service
}

func New(deps Dependencies) *Server {
	return &Server{
		inventory: service.New(deps.Pvz, deps.ActiveVacancies),
	}
}

func (s *Server) CreatePvz(ctx context.Context, req *inventoryv1.CreatePvzRequest) (*inventoryv1.CreatePvzResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	pvz, err := s.inventory.CreatePvz(ctx, dto.CreatePvzInput{
		Actor:           actor,
		City:            req.GetCity(),
		Address:         req.GetAddress(),
		CompanyName:     req.GetCompanyName(),
		ContactPhone:    req.GetContactPhone(),
		ContactTelegram: req.ContactTelegram,
		Description:     req.Description,
		Lat:             req.Lat,
		Lng:             req.Lng,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &inventoryv1.CreatePvzResponse{Pvz: toProtoPvz(pvz)}, nil
}

func (s *Server) GetPvz(ctx context.Context, req *inventoryv1.GetPvzRequest) (*inventoryv1.GetPvzResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	pvz, err := s.inventory.GetPvz(ctx, dto.GetPvzInput{Actor: actor, ID: req.GetId()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &inventoryv1.GetPvzResponse{Pvz: toProtoPvz(pvz)}, nil
}

func (s *Server) ListOwnerPvz(ctx context.Context, req *inventoryv1.ListOwnerPvzRequest) (*inventoryv1.ListOwnerPvzResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	items, next, err := s.inventory.ListOwnerPvz(ctx, dto.ListOwnerPvzInput{
		Actor:     actor,
		PageSize:  req.GetPageSize(),
		PageToken: req.PageToken,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &inventoryv1.ListOwnerPvzResponse{Items: toProtoPvzList(items), NextPageToken: next}, nil
}

func (s *Server) UpdatePvz(ctx context.Context, req *inventoryv1.UpdatePvzRequest) (*inventoryv1.UpdatePvzResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	pvz, err := s.inventory.UpdatePvz(ctx, dto.UpdatePvzInput{
		Actor:           actor,
		ID:              req.GetId(),
		City:            req.City,
		Address:         req.Address,
		CompanyName:     req.CompanyName,
		ContactPhone:    req.ContactPhone,
		ContactTelegram: req.ContactTelegram,
		Description:     req.Description,
		Lat:             req.Lat,
		Lng:             req.Lng,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &inventoryv1.UpdatePvzResponse{Pvz: toProtoPvz(pvz)}, nil
}

func (s *Server) CheckPvzForVacancy(ctx context.Context, req *inventoryv1.CheckPvzForVacancyRequest) (*inventoryv1.CheckPvzForVacancyResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	pvz, err := s.inventory.CheckPvzForVacancy(ctx, dto.CheckPvzForVacancyInput{Actor: actor, PvzID: req.GetPvzId()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &inventoryv1.CheckPvzForVacancyResponse{
		Pvz: &inventoryv1.VacancyPvzSnapshot{
			PvzId:       pvz.ID,
			OwnerId:     pvz.OwnerID,
			CompanyName: pvz.CompanyName,
			Address:     pvz.Address,
		},
	}, nil
}

func (s *Server) ListPendingPvz(ctx context.Context, req *inventoryv1.ListPendingPvzRequest) (*inventoryv1.ListPendingPvzResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	items, next, err := s.inventory.ListPendingPvz(ctx, dto.ListPendingPvzInput{
		Actor:     actor,
		PageSize:  req.GetPageSize(),
		PageToken: req.PageToken,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &inventoryv1.ListPendingPvzResponse{Items: toProtoPvzList(items), NextPageToken: next}, nil
}

func (s *Server) ApprovePvz(ctx context.Context, req *inventoryv1.ApprovePvzRequest) (*inventoryv1.ApprovePvzResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	pvz, err := s.inventory.ApprovePvz(ctx, dto.ModeratePvzInput{Actor: actor, PvzID: req.GetPvzId(), Comment: req.Comment})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &inventoryv1.ApprovePvzResponse{Pvz: toProtoPvz(pvz)}, nil
}

func (s *Server) RejectPvz(ctx context.Context, req *inventoryv1.RejectPvzRequest) (*inventoryv1.RejectPvzResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	reason := req.GetReason()
	pvz, err := s.inventory.RejectPvz(ctx, dto.ModeratePvzInput{Actor: actor, PvzID: req.GetPvzId(), Reason: &reason, Comment: req.Comment})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &inventoryv1.RejectPvzResponse{Pvz: toProtoPvz(pvz)}, nil
}

func (s *Server) BlockPvz(ctx context.Context, req *inventoryv1.BlockPvzRequest) (*inventoryv1.BlockPvzResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	pvz, err := s.inventory.BlockPvz(ctx, dto.ModeratePvzInput{Actor: actor, PvzID: req.GetPvzId(), Comment: req.Comment})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &inventoryv1.BlockPvzResponse{Pvz: toProtoPvz(pvz)}, nil
}

func (s *Server) DeleteOwnerPvzRequests(ctx context.Context, req *inventoryv1.DeleteOwnerPvzRequestsRequest) (*inventoryv1.DeleteOwnerPvzRequestsResponse, error) {
	actor, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	deleted, blocked, err := s.inventory.DeleteOwnerPvzRequests(ctx, dto.DeleteOwnerPvzRequestsInput{
		Actor:   actor,
		OwnerID: req.GetOwnerId(),
		Comment: req.Comment,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &inventoryv1.DeleteOwnerPvzRequestsResponse{DeletedCount: deleted, BlockedCount: blocked}, nil
}

func actorFromMetadata(ctx context.Context) (dto.Actor, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return dto.Actor{}, status.Error(codes.Unauthenticated, "missing identity metadata")
	}
	ids := md.Get("x-user-id")
	if len(ids) == 0 {
		return dto.Actor{}, status.Error(codes.Unauthenticated, "missing identity metadata")
	}
	userID, err := strconv.ParseInt(ids[0], 10, 64)
	if err != nil || userID <= 0 {
		return dto.Actor{}, status.Error(codes.Unauthenticated, "invalid identity metadata")
	}

	roles := md.Get("x-user-role")
	if len(roles) == 0 {
		return dto.Actor{}, status.Error(codes.Unauthenticated, "missing identity metadata")
	}
	role := roleFromString(roles[0])
	if role == commonv1.UserRole_USER_ROLE_UNSPECIFIED {
		return dto.Actor{}, status.Error(codes.Unauthenticated, "invalid identity metadata")
	}

	return dto.Actor{UserID: userID, Role: role}, nil
}

func toProtoPvz(pvz store.Pvz) *inventoryv1.Pvz {
	out := &inventoryv1.Pvz{
		Id:              pvz.ID,
		OwnerId:         pvz.OwnerID,
		Status:          pvz.Status,
		City:            pvz.City,
		Address:         pvz.Address,
		CompanyName:     pvz.CompanyName,
		ContactPhone:    pvz.ContactPhone,
		ContactTelegram: pvz.ContactTelegram,
		Description:     pvz.Description,
		Lat:             pvz.Lat,
		Lng:             pvz.Lng,
		RejectionReason: pvz.RejectionReason,
		ModeratedBy:     pvz.ModeratedBy,
		CreatedAt:       timestamppb.New(pvz.CreatedAt),
		UpdatedAt:       timestamppb.New(pvz.UpdatedAt),
	}
	if pvz.ModeratedAt != nil {
		out.ModeratedAt = timestamppb.New(*pvz.ModeratedAt)
	}
	return out
}

func toProtoPvzList(items []store.Pvz) []*inventoryv1.Pvz {
	out := make([]*inventoryv1.Pvz, 0, len(items))
	for _, item := range items {
		out = append(out, toProtoPvz(item))
	}
	return out
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
	case errors.Is(svcErr.Kind, service.ErrUnauthenticated):
		code = codes.Unauthenticated
	case errors.Is(svcErr.Kind, service.ErrPermissionDenied):
		code = codes.PermissionDenied
	case errors.Is(svcErr.Kind, service.ErrNotFound):
		code = codes.NotFound
	case errors.Is(svcErr.Kind, service.ErrUnavailable):
		code = codes.Unavailable
	case errors.Is(svcErr.Kind, service.ErrFailedPrecondition):
		code = codes.FailedPrecondition
	}
	return status.Error(code, svcErr.Message)
}
