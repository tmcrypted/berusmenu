package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gitlab.com/berusmenu/auth-service/internal/config"
	"gitlab.com/berusmenu/auth-service/internal/events"
	"gitlab.com/berusmenu/auth-service/internal/store"
	"gitlab.com/berusmenu/auth-service/internal/verification"
	authv1 "gitlab.com/berusmenu/proto/gen/go/auth/v1"
	commonv1 "gitlab.com/berusmenu/proto/gen/go/common/v1"
	"golang.org/x/crypto/bcrypt"
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
	cfg          config.Config
	users        *store.Store
	verification *verification.Store
	publisher    *events.Publisher
}

type accessClaims struct {
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	jwt.RegisteredClaims
}

func New(deps Dependencies) *Server {
	return &Server{
		cfg:          deps.Config,
		users:        deps.Users,
		verification: deps.Verification,
		publisher:    deps.Publisher,
	}
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	email := store.NormalizeEmail(req.GetEmail())
	if email == "" || req.GetPassword() == "" || strings.TrimSpace(req.GetFullName()) == "" {
		return nil, status.Error(codes.InvalidArgument, "email, password and full_name are required")
	}
	if len(req.GetPassword()) < 8 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}
	role := req.GetRole()
	if role == commonv1.UserRole_USER_ROLE_UNSPECIFIED {
		role = commonv1.UserRole_USER_ROLE_WORKER
	}
	if role != commonv1.UserRole_USER_ROLE_OWNER && role != commonv1.UserRole_USER_ROLE_WORKER && role != commonv1.UserRole_USER_ROLE_MODERATOR {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	refreshToken, err := randomToken()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}
	refreshExpiresAt := time.Now().Add(s.cfg.RefreshTTL)

	user, err := s.users.CreateUserWithSession(ctx, store.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         role,
		FullName:     strings.TrimSpace(req.GetFullName()),
		Phone:        req.Phone,
	}, refreshToken, refreshExpiresAt)
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "email already registered")
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	codeState, err := s.verification.CreateOrReuseCode(ctx, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create verification code")
	}
	if err := s.publisher.EmailVerificationRequested(ctx, user.ID, user.Email, codeState.Code); err != nil {
		return nil, status.Error(codes.Unavailable, "failed to publish verification email event")
	}

	auth, err := s.authPayload(user, refreshToken)
	if err != nil {
		return nil, err
	}
	return &authv1.RegisterResponse{Auth: auth}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	user, err := s.users.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	if user.Status != "active" {
		return nil, status.Error(codes.PermissionDenied, "user is blocked")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.GetPassword())); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	refreshToken, err := randomToken()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}
	if err := s.users.CreateRefreshSession(ctx, user.ID, refreshToken, time.Now().Add(s.cfg.RefreshTTL)); err != nil {
		return nil, status.Error(codes.Internal, "failed to create refresh session")
	}

	auth, err := s.authPayload(user, refreshToken)
	if err != nil {
		return nil, err
	}
	return &authv1.LoginResponse{Auth: auth}, nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	user, err := s.users.RefreshSessionUser(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	if user.Status != "active" {
		return nil, status.Error(codes.PermissionDenied, "user is blocked")
	}
	if err := s.users.RevokeRefreshSession(ctx, req.GetRefreshToken()); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	refreshToken, err := randomToken()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}
	if err := s.users.CreateRefreshSession(ctx, user.ID, refreshToken, time.Now().Add(s.cfg.RefreshTTL)); err != nil {
		return nil, status.Error(codes.Internal, "failed to create refresh session")
	}

	auth, err := s.authPayload(user, refreshToken)
	if err != nil {
		return nil, err
	}
	return &authv1.RefreshResponse{Auth: auth}, nil
}

func (s *Server) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := s.users.RevokeRefreshSession(ctx, req.GetRefreshToken()); err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, status.Error(codes.Internal, "failed to revoke refresh token")
	}
	return &authv1.LogoutResponse{}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, err := s.parseAccessToken(req.GetAccessToken())
	if err != nil {
		return &authv1.ValidateTokenResponse{Valid: false}, nil
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return &authv1.ValidateTokenResponse{Valid: false}, nil
	}
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil || user.Status != "active" {
		return &authv1.ValidateTokenResponse{Valid: false}, nil
	}

	return &authv1.ValidateTokenResponse{
		Valid:         true,
		UserId:        user.ID,
		Role:          user.Role,
		Email:         user.Email,
		EmailVerified: user.EmailVerifiedAt != nil,
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.GetUserResponse, error) {
	if err := requireSameUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}
	user, err := s.users.GetUserByID(ctx, req.GetUserId())
	if err != nil {
		return nil, notFoundOrInternal(err, "user not found", "failed to get user")
	}
	return &authv1.GetUserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *authv1.UpdateProfileRequest) (*authv1.UpdateProfileResponse, error) {
	userID, err := userIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.users.UpdateProfile(ctx, userID, req.FullName, req.Phone, req.AvatarUrl)
	if err != nil {
		return nil, notFoundOrInternal(err, "user not found", "failed to update profile")
	}
	return &authv1.UpdateProfileResponse{User: toProtoUser(user)}, nil
}

func (s *Server) ResendVerificationCode(ctx context.Context, req *authv1.ResendVerificationCodeRequest) (*authv1.ResendVerificationCodeResponse, error) {
	userID, err := userIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, notFoundOrInternal(err, "user not found", "failed to get user")
	}
	if user.EmailVerifiedAt != nil {
		return nil, status.Error(codes.FailedPrecondition, "email already verified")
	}

	codeState, err := s.verification.ResendCode(ctx, userID)
	if err != nil {
		if errors.Is(err, verification.ErrCooldown) {
			return nil, status.Error(codes.ResourceExhausted, "resend cooldown is active")
		}
		if errors.Is(err, verification.ErrLockout) {
			return nil, status.Error(codes.ResourceExhausted, "verification attempts lockout is active")
		}
		return nil, status.Error(codes.Internal, "failed to create verification code")
	}
	if err := s.publisher.EmailVerificationRequested(ctx, user.ID, user.Email, codeState.Code); err != nil {
		return nil, status.Error(codes.Unavailable, "failed to publish verification email event")
	}

	return &authv1.ResendVerificationCodeResponse{
		CodeExpiresAt:     timestamppb.New(codeState.CodeExpiresAt),
		ResendAvailableAt: timestamppb.New(codeState.ResendAvailableAt),
	}, nil
}

func (s *Server) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	userID, err := userIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.verification.VerifyCode(ctx, userID, strings.TrimSpace(req.GetCode())); err != nil {
		if errors.Is(err, verification.ErrBadCode) || errors.Is(err, verification.ErrNoCode) {
			return nil, status.Error(codes.InvalidArgument, "invalid verification code")
		}
		if errors.Is(err, verification.ErrLockout) {
			return nil, status.Error(codes.ResourceExhausted, "verification attempts lockout is active")
		}
		return nil, status.Error(codes.Internal, "failed to verify code")
	}
	user, err := s.users.MarkEmailVerified(ctx, userID)
	if err != nil {
		return nil, notFoundOrInternal(err, "user not found", "failed to verify email")
	}
	return &authv1.VerifyEmailResponse{User: toProtoUser(user)}, nil
}

func (s *Server) authPayload(user store.User, refreshToken string) (*authv1.AuthPayload, error) {
	accessToken, accessExpiresAt, err := s.issueAccessToken(user)
	if err != nil {
		return nil, err
	}
	return &authv1.AuthPayload{
		Tokens: &authv1.AuthTokens{
			AccessToken:     accessToken,
			RefreshToken:    refreshToken,
			AccessExpiresAt: timestamppb.New(accessExpiresAt),
		},
		User: toProtoUser(user),
	}, nil
}

func (s *Server) issueAccessToken(user store.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.cfg.AccessTTL)
	claims := accessClaims{
		Email:         user.Email,
		Role:          roleString(user.Role),
		EmailVerified: user.EmailVerifiedAt != nil,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, status.Error(codes.Internal, "failed to sign access token")
	}
	return token, expiresAt, nil
}

func (s *Server) parseAccessToken(raw string) (*accessClaims, error) {
	claims := &accessClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func toProtoUser(user store.User) *authv1.User {
	out := &authv1.User{
		Id:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		FullName:  user.FullName,
		Phone:     user.Phone,
		AvatarUrl: user.AvatarURL,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
	if user.EmailVerifiedAt != nil {
		out.EmailVerifiedAt = timestamppb.New(*user.EmailVerifiedAt)
	}
	return out
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
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

func requireSameUser(ctx context.Context, requestedID int64) error {
	userID, err := userIDFromMetadata(ctx)
	if err != nil {
		return err
	}
	if userID != requestedID {
		return status.Error(codes.PermissionDenied, "cannot access another user")
	}
	return nil
}

func notFoundOrInternal(err error, notFoundMessage string, internalMessage string) error {
	if errors.Is(err, store.ErrNotFound) {
		return status.Error(codes.NotFound, notFoundMessage)
	}
	return status.Error(codes.Internal, internalMessage)
}

func roleString(role commonv1.UserRole) string {
	switch role {
	case commonv1.UserRole_USER_ROLE_OWNER:
		return "owner"
	case commonv1.UserRole_USER_ROLE_MODERATOR:
		return "moderator"
	default:
		return "worker"
	}
}
