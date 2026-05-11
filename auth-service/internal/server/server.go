package server

import (
	"context"
	"errors"
	"strconv"

	"github.com/tmcrypted/berusmenu/auth-service/internal/config"
	"github.com/tmcrypted/berusmenu/auth-service/internal/dto"
	"github.com/tmcrypted/berusmenu/auth-service/internal/events"
	"github.com/tmcrypted/berusmenu/auth-service/internal/service"
	"github.com/tmcrypted/berusmenu/auth-service/internal/store"
	"github.com/tmcrypted/berusmenu/auth-service/internal/verification"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Dependencies struct {
	Config       config.Config
	Users        *store.Store
	Verification *verification.Store
	Publisher    *events.Publisher
}

type Server struct {
	authv1.UnimplementedAuthServiceServer
	auth *service.Service
}

func New(deps Dependencies) *Server {
	return &Server{
		auth: service.New(deps.Config, deps.Users, deps.Verification, deps.Publisher),
	}
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	auth, err := s.auth.Register(ctx, dto.RegisterInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Role:     req.GetRole(),
		Phone:    req.Phone,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.RegisterResponse{Auth: toProtoAuthPayload(auth)}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	auth, err := s.auth.Login(ctx, dto.LoginInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.LoginResponse{Auth: toProtoAuthPayload(auth)}, nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	auth, err := s.auth.Refresh(ctx, dto.RefreshInput{RefreshToken: req.GetRefreshToken()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.RefreshResponse{Auth: toProtoAuthPayload(auth)}, nil
}

func (s *Server) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := s.auth.Logout(ctx, dto.LogoutInput{RefreshToken: req.GetRefreshToken()}); err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.LogoutResponse{}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	out, err := s.auth.ValidateToken(ctx, dto.ValidateTokenInput{AccessToken: req.GetAccessToken()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.ValidateTokenResponse{
		Valid:         out.Valid,
		UserId:        out.UserID,
		Role:          out.Role,
		Email:         out.Email,
		EmailVerified: out.EmailVerified,
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.GetUserResponse, error) {
	userID, err := userIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.auth.GetUser(ctx, dto.GetUserInput{
		CurrentUserID: userID,
		UserID:        req.GetUserId(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.GetUserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) GetUserByID(ctx context.Context, req *authv1.GetUserByIDRequest) (*authv1.GetUserByIDResponse, error) {
	if _, err := userIDFromMetadata(ctx); err != nil {
		return nil, err
	}
	user, err := s.auth.GetUserByID(ctx, req.GetUserId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.GetUserByIDResponse{User: toProtoUser(user)}, nil
}

func (s *Server) BlockUser(ctx context.Context, req *authv1.BlockUserRequest) (*authv1.BlockUserResponse, error) {
	actorUserID, actorRole, err := actorFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetBlockedUntil() == nil {
		return nil, status.Error(codes.InvalidArgument, "blocked_until is required")
	}
	user, err := s.auth.BlockUser(ctx, dto.BlockUserInput{
		ActorUserID:  actorUserID,
		ActorRole:    actorRole,
		UserID:       req.GetUserId(),
		Reason:       req.Reason,
		BlockedUntil: req.GetBlockedUntil().AsTime(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.BlockUserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *authv1.UpdateProfileRequest) (*authv1.UpdateProfileResponse, error) {
	userID, err := userIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.auth.UpdateProfile(ctx, dto.UpdateProfileInput{
		UserID:           userID,
		LastName:         req.LastName,
		FirstName:        req.FirstName,
		MiddleName:       req.MiddleName,
		BirthDate:        req.BirthDate,
		Phone:            req.Phone,
		AvatarURL:        req.AvatarUrl,
		TelegramUsername: req.TelegramUsername,
		MaxUsername:      req.MaxUsername,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.UpdateProfileResponse{User: toProtoUser(user)}, nil
}

func (s *Server) ResendVerificationCode(ctx context.Context, req *authv1.ResendVerificationCodeRequest) (*authv1.ResendVerificationCodeResponse, error) {
	userID, err := userIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	out, err := s.auth.ResendVerificationCode(ctx, dto.ResendVerificationCodeInput{UserID: userID})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.ResendVerificationCodeResponse{
		CodeExpiresAt:     timestamppb.New(out.CodeExpiresAt),
		ResendAvailableAt: timestamppb.New(out.ResendAvailableAt),
	}, nil
}

func (s *Server) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	userID, err := userIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.auth.VerifyEmail(ctx, dto.VerifyEmailInput{
		UserID: userID,
		Code:   req.GetCode(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.VerifyEmailResponse{User: toProtoUser(user)}, nil
}

func toProtoAuthPayload(auth dto.AuthPayload) *authv1.AuthPayload {
	return &authv1.AuthPayload{
		Tokens: &authv1.AuthTokens{
			AccessToken:     auth.AccessToken,
			RefreshToken:    auth.RefreshToken,
			AccessExpiresAt: timestamppb.New(auth.AccessExpiresAt),
		},
		User: toProtoUser(auth.User),
	}
}

func toProtoUser(user store.User) *authv1.User {
	out := &authv1.User{
		Id:               user.ID,
		Email:            user.Email,
		Role:             user.Role,
		LastName:         user.LastName,
		FirstName:        user.FirstName,
		Phone:            user.Phone,
		AvatarUrl:        user.AvatarURL,
		TelegramUsername: user.TelegramUsername,
		MaxUsername:      user.MaxUsername,
		CreatedAt:        timestamppb.New(user.CreatedAt),
		UpdatedAt:        timestamppb.New(user.UpdatedAt),
	}
	if user.MiddleName != nil {
		out.MiddleName = user.MiddleName
	}
	if user.BirthDate != nil {
		value := user.BirthDate.Format("2006-01-02")
		out.BirthDate = &value
	}
	if user.EmailVerifiedAt != nil {
		out.EmailVerifiedAt = timestamppb.New(*user.EmailVerifiedAt)
	}
	return out
}

func userIDFromMetadata(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing identity metadata")
	}
	values := md.Get("x-user-id")
	if len(values) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing identity metadata")
	}
	userID, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || userID <= 0 {
		return 0, status.Error(codes.Unauthenticated, "invalid identity metadata")
	}
	return userID, nil
}

func actorFromMetadata(ctx context.Context) (int64, commonv1.UserRole, error) {
	userID, err := userIDFromMetadata(ctx)
	if err != nil {
		return 0, commonv1.UserRole_USER_ROLE_UNSPECIFIED, err
	}

	md, _ := metadata.FromIncomingContext(ctx)
	roles := md.Get("x-user-role")
	if len(roles) == 0 {
		return 0, commonv1.UserRole_USER_ROLE_UNSPECIFIED, status.Error(codes.Unauthenticated, "missing identity metadata")
	}
	role := roleFromMetadata(roles[0])
	if role == commonv1.UserRole_USER_ROLE_UNSPECIFIED {
		return 0, commonv1.UserRole_USER_ROLE_UNSPECIFIED, status.Error(codes.Unauthenticated, "invalid identity metadata")
	}
	return userID, role, nil
}

func roleFromMetadata(value string) commonv1.UserRole {
	switch value {
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
	case errors.Is(svcErr.Kind, service.ErrUnavailable):
		code = codes.Unavailable
	case errors.Is(svcErr.Kind, service.ErrFailedPrecondition):
		code = codes.FailedPrecondition
	case errors.Is(svcErr.Kind, service.ErrResourceExhausted):
		code = codes.ResourceExhausted
	}
	return status.Error(code, svcErr.Message)
}
