package service

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
	"github.com/tmcrypted/berusmenu/auth-service/internal/config"
	"github.com/tmcrypted/berusmenu/auth-service/internal/dto"
	"github.com/tmcrypted/berusmenu/auth-service/internal/store"
	"github.com/tmcrypted/berusmenu/auth-service/internal/verification"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUserWithSession(ctx context.Context, user store.User, refreshToken string, refreshExpiresAt time.Time) (store.User, error)
	GetUserByEmail(ctx context.Context, email string) (store.User, error)
	GetUserByID(ctx context.Context, id int64) (store.User, error)
	CreateRefreshSession(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error
	RefreshSessionUser(ctx context.Context, refreshToken string) (store.User, error)
	RotateRefreshSession(ctx context.Context, oldRefreshToken string, newRefreshToken string, newExpiresAt time.Time) (store.User, error)
	RevokeRefreshSession(ctx context.Context, refreshToken string) error
	MarkEmailVerified(ctx context.Context, userID int64) (store.User, error)
	UpdateProfile(ctx context.Context, userID int64, fullName *string, phone *string, avatarURL *string) (store.User, error)
}

type VerificationStore interface {
	CreateOrReuseCode(ctx context.Context, userID int64) (verification.CodeState, error)
	ResendCode(ctx context.Context, userID int64) (verification.CodeState, error)
	VerifyCode(ctx context.Context, userID int64, submitted string) error
}

type EventPublisher interface {
	EmailVerificationRequested(ctx context.Context, userID int64, email string, code string) error
}

type Service struct {
	cfg          config.Config
	users        UserRepository
	verification VerificationStore
	publisher    EventPublisher
}

type accessClaims struct {
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	jwt.RegisteredClaims
}

func New(cfg config.Config, users UserRepository, verification VerificationStore, publisher EventPublisher) *Service {
	return &Service{
		cfg:          cfg,
		users:        users,
		verification: verification,
		publisher:    publisher,
	}
}

func (s *Service) Register(ctx context.Context, input dto.RegisterInput) (dto.AuthPayload, error) {
	email := store.NormalizeEmail(input.Email)
	if email == "" || input.Password == "" || strings.TrimSpace(input.FullName) == "" {
		return dto.AuthPayload{}, serviceError(ErrInvalidArgument, "email, password and full_name are required")
	}
	if len(input.Password) < 8 {
		return dto.AuthPayload{}, serviceError(ErrInvalidArgument, "password must be at least 8 characters")
	}
	role := input.Role
	if role == commonv1.UserRole_USER_ROLE_UNSPECIFIED {
		role = commonv1.UserRole_USER_ROLE_WORKER
	}
	if role != commonv1.UserRole_USER_ROLE_OWNER && role != commonv1.UserRole_USER_ROLE_WORKER && role != commonv1.UserRole_USER_ROLE_MODERATOR {
		return dto.AuthPayload{}, serviceError(ErrInvalidArgument, "invalid role")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.AuthPayload{}, serviceError(ErrInternal, "failed to hash password")
	}

	refreshToken, err := randomToken()
	if err != nil {
		return dto.AuthPayload{}, serviceError(ErrInternal, "failed to generate refresh token")
	}
	user, err := s.users.CreateUserWithSession(ctx, store.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         role,
		FullName:     strings.TrimSpace(input.FullName),
		Phone:        input.Phone,
	}, refreshToken, time.Now().Add(s.cfg.RefreshTTL))
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return dto.AuthPayload{}, serviceError(ErrAlreadyExists, "email already registered")
		}
		return dto.AuthPayload{}, serviceError(ErrInternal, "failed to create user")
	}

	codeState, err := s.verification.CreateOrReuseCode(ctx, user.ID)
	if err != nil {
		return dto.AuthPayload{}, serviceError(ErrInternal, "failed to create verification code")
	}
	if err := s.publisher.EmailVerificationRequested(ctx, user.ID, user.Email, codeState.Code); err != nil {
		return dto.AuthPayload{}, serviceError(ErrUnavailable, "failed to publish verification email event")
	}

	return s.authPayload(user, refreshToken)
}

func (s *Service) Login(ctx context.Context, input dto.LoginInput) (dto.AuthPayload, error) {
	user, err := s.users.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return dto.AuthPayload{}, serviceError(ErrUnauthenticated, "invalid email or password")
	}
	if user.Status != "active" {
		return dto.AuthPayload{}, serviceError(ErrPermissionDenied, "user is blocked")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return dto.AuthPayload{}, serviceError(ErrUnauthenticated, "invalid email or password")
	}

	refreshToken, err := randomToken()
	if err != nil {
		return dto.AuthPayload{}, serviceError(ErrInternal, "failed to generate refresh token")
	}
	if err := s.users.CreateRefreshSession(ctx, user.ID, refreshToken, time.Now().Add(s.cfg.RefreshTTL)); err != nil {
		return dto.AuthPayload{}, serviceError(ErrInternal, "failed to create refresh session")
	}

	return s.authPayload(user, refreshToken)
}

func (s *Service) Refresh(ctx context.Context, input dto.RefreshInput) (dto.AuthPayload, error) {
	refreshToken, err := randomToken()
	if err != nil {
		return dto.AuthPayload{}, serviceError(ErrInternal, "failed to generate refresh token")
	}
	user, err := s.users.RotateRefreshSession(ctx, input.RefreshToken, refreshToken, time.Now().Add(s.cfg.RefreshTTL))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return dto.AuthPayload{}, serviceError(ErrUnauthenticated, "invalid refresh token")
		}
		return dto.AuthPayload{}, serviceError(ErrInternal, "failed to rotate refresh session")
	}
	if user.Status != "active" {
		return dto.AuthPayload{}, serviceError(ErrPermissionDenied, "user is blocked")
	}

	return s.authPayload(user, refreshToken)
}

func (s *Service) Logout(ctx context.Context, input dto.LogoutInput) error {
	if err := s.users.RevokeRefreshSession(ctx, input.RefreshToken); err != nil && !errors.Is(err, store.ErrNotFound) {
		return serviceError(ErrInternal, "failed to revoke refresh token")
	}
	return nil
}

func (s *Service) ValidateToken(ctx context.Context, input dto.ValidateTokenInput) (dto.ValidateTokenOutput, error) {
	claims, err := s.parseAccessToken(input.AccessToken)
	if err != nil {
		return dto.ValidateTokenOutput{Valid: false}, nil
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return dto.ValidateTokenOutput{Valid: false}, nil
	}
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil || user.Status != "active" {
		return dto.ValidateTokenOutput{Valid: false}, nil
	}

	return dto.ValidateTokenOutput{
		Valid:         true,
		UserID:        user.ID,
		Role:          user.Role,
		Email:         user.Email,
		EmailVerified: user.EmailVerifiedAt != nil,
	}, nil
}

func (s *Service) GetUser(ctx context.Context, input dto.GetUserInput) (store.User, error) {
	if input.CurrentUserID != input.UserID {
		return store.User{}, serviceError(ErrPermissionDenied, "cannot access another user")
	}
	user, err := s.users.GetUserByID(ctx, input.UserID)
	if err != nil {
		return store.User{}, notFoundOrInternal(err, "user not found", "failed to get user")
	}
	return user, nil
}

func (s *Service) UpdateProfile(ctx context.Context, input dto.UpdateProfileInput) (store.User, error) {
	user, err := s.users.UpdateProfile(ctx, input.UserID, input.FullName, input.Phone, input.AvatarURL)
	if err != nil {
		return store.User{}, notFoundOrInternal(err, "user not found", "failed to update profile")
	}
	return user, nil
}

func (s *Service) ResendVerificationCode(ctx context.Context, input dto.ResendVerificationCodeInput) (dto.VerificationCodeOutput, error) {
	user, err := s.users.GetUserByID(ctx, input.UserID)
	if err != nil {
		return dto.VerificationCodeOutput{}, notFoundOrInternal(err, "user not found", "failed to get user")
	}
	if user.EmailVerifiedAt != nil {
		return dto.VerificationCodeOutput{}, serviceError(ErrFailedPrecondition, "email already verified")
	}

	codeState, err := s.verification.ResendCode(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, verification.ErrCooldown) {
			return dto.VerificationCodeOutput{}, serviceError(ErrResourceExhausted, "resend cooldown is active")
		}
		if errors.Is(err, verification.ErrLockout) {
			return dto.VerificationCodeOutput{}, serviceError(ErrResourceExhausted, "verification attempts lockout is active")
		}
		return dto.VerificationCodeOutput{}, serviceError(ErrInternal, "failed to create verification code")
	}
	if err := s.publisher.EmailVerificationRequested(ctx, user.ID, user.Email, codeState.Code); err != nil {
		return dto.VerificationCodeOutput{}, serviceError(ErrUnavailable, "failed to publish verification email event")
	}

	return dto.VerificationCodeOutput{
		CodeExpiresAt:     codeState.CodeExpiresAt,
		ResendAvailableAt: codeState.ResendAvailableAt,
	}, nil
}

func (s *Service) VerifyEmail(ctx context.Context, input dto.VerifyEmailInput) (store.User, error) {
	if err := s.verification.VerifyCode(ctx, input.UserID, strings.TrimSpace(input.Code)); err != nil {
		if errors.Is(err, verification.ErrBadCode) || errors.Is(err, verification.ErrNoCode) {
			return store.User{}, serviceError(ErrInvalidArgument, "invalid verification code")
		}
		if errors.Is(err, verification.ErrLockout) {
			return store.User{}, serviceError(ErrResourceExhausted, "verification attempts lockout is active")
		}
		return store.User{}, serviceError(ErrInternal, "failed to verify code")
	}
	user, err := s.users.MarkEmailVerified(ctx, input.UserID)
	if err != nil {
		return store.User{}, notFoundOrInternal(err, "user not found", "failed to verify email")
	}
	return user, nil
}

func (s *Service) authPayload(user store.User, refreshToken string) (dto.AuthPayload, error) {
	accessToken, accessExpiresAt, err := s.issueAccessToken(user)
	if err != nil {
		return dto.AuthPayload{}, err
	}
	return dto.AuthPayload{
		User:            user,
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		AccessExpiresAt: accessExpiresAt,
	}, nil
}

func (s *Service) issueAccessToken(user store.User) (string, time.Time, error) {
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
		return "", time.Time{}, serviceError(ErrInternal, "failed to sign access token")
	}
	return token, expiresAt, nil
}

func (s *Service) parseAccessToken(raw string) (*accessClaims, error) {
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

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func notFoundOrInternal(err error, notFoundMessage string, internalMessage string) error {
	if errors.Is(err, store.ErrNotFound) {
		return serviceError(ErrNotFound, notFoundMessage)
	}
	return serviceError(ErrInternal, internalMessage)
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
