package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tmcrypted/berusmenu/auth-service/internal/config"
	"github.com/tmcrypted/berusmenu/auth-service/internal/dto"
	"github.com/tmcrypted/berusmenu/auth-service/internal/store"
	"github.com/tmcrypted/berusmenu/auth-service/internal/verification"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		name    string
		input   dto.RegisterInput
		setup   func(*MockUserRepository, *MockVerificationStore, *MockEventPublisher)
		wantErr error
	}{
		{
			name: "success",
			input: dto.RegisterInput{
				Email:    " OWNER@EXAMPLE.COM ",
				Password: "password123",
				Role:     commonv1.UserRole_USER_ROLE_OWNER,
			},
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore, publisher *MockEventPublisher) {
				users.EXPECT().
					CreateUserWithSession(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user store.User, refreshToken string, expiresAt time.Time) (store.User, error) {
						if user.Email != "owner@example.com" {
							t.Fatalf("email = %q", user.Email)
						}
						if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
							t.Fatalf("password hash does not match: %v", err)
						}
						user.ID = 7
						user.Status = "active"
						user.CreatedAt = time.Now()
						user.UpdatedAt = time.Now()
						return user, nil
					})
				verificationStore.EXPECT().CreateOrReuseCode(gomock.Any(), int64(7)).Return(verification.CodeState{Code: "123456"}, nil)
				publisher.EXPECT().EmailVerificationRequested(gomock.Any(), int64(7), "owner@example.com", "123456").Return(nil)
			},
		},
		{
			name: "duplicate email",
			input: dto.RegisterInput{
				Email:    "owner@example.com",
				Password: "password123",
			},
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore, publisher *MockEventPublisher) {
				users.EXPECT().CreateUserWithSession(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(store.User{}, store.ErrAlreadyExists)
			},
			wantErr: ErrAlreadyExists,
		},
		{
			name: "invalid input",
			input: dto.RegisterInput{
				Email:    "",
				Password: "short",
			},
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore, publisher *MockEventPublisher) {
			},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "publisher failure",
			input: dto.RegisterInput{
				Email:    "owner@example.com",
				Password: "password123",
			},
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore, publisher *MockEventPublisher) {
				user := sampleUser()
				users.EXPECT().CreateUserWithSession(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(user, nil)
				verificationStore.EXPECT().CreateOrReuseCode(gomock.Any(), user.ID).Return(verification.CodeState{Code: "123456"}, nil)
				publisher.EXPECT().EmailVerificationRequested(gomock.Any(), user.ID, user.Email, "123456").Return(errors.New("kafka down"))
			},
			wantErr: ErrUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			users := NewMockUserRepository(ctrl)
			verificationStore := NewMockVerificationStore(ctrl)
			publisher := NewMockEventPublisher(ctrl)
			tt.setup(users, verificationStore, publisher)

			svc := New(testConfig(), users, verificationStore, publisher)
			_, err := svc.Register(context.Background(), tt.input)
			assertErrorKind(t, err, tt.wantErr)
		})
	}
}

func TestLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		setup   func(*MockUserRepository)
		wantErr error
	}{
		{
			name: "success",
			setup: func(users *MockUserRepository) {
				user := sampleUser()
				user.PasswordHash = string(hash)
				users.EXPECT().GetUserByEmail(gomock.Any(), "worker@example.com").Return(user, nil)
				users.EXPECT().CreateRefreshSession(gomock.Any(), user.ID, gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "wrong password",
			setup: func(users *MockUserRepository) {
				user := sampleUser()
				user.PasswordHash = string(hash)
				users.EXPECT().GetUserByEmail(gomock.Any(), "worker@example.com").Return(user, nil)
			},
			wantErr: ErrUnauthenticated,
		},
		{
			name: "blocked user",
			setup: func(users *MockUserRepository) {
				user := sampleUser()
				user.PasswordHash = string(hash)
				user.Status = "blocked"
				users.EXPECT().GetUserByEmail(gomock.Any(), "worker@example.com").Return(user, nil)
			},
			wantErr: ErrPermissionDenied,
		},
		{
			name: "expired block can login",
			setup: func(users *MockUserRepository) {
				user := sampleUser()
				user.PasswordHash = string(hash)
				user.Status = "blocked"
				expired := time.Now().Add(-time.Minute)
				user.BlockedUntil = &expired
				users.EXPECT().GetUserByEmail(gomock.Any(), "worker@example.com").Return(user, nil)
				users.EXPECT().CreateRefreshSession(gomock.Any(), user.ID, gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			users := NewMockUserRepository(ctrl)
			tt.setup(users)

			password := "password123"
			if tt.name == "wrong password" {
				password = "wrong-password"
			}
			svc := New(testConfig(), users, NewMockVerificationStore(ctrl), NewMockEventPublisher(ctrl))
			_, err := svc.Login(context.Background(), dto.LoginInput{Email: "worker@example.com", Password: password})
			assertErrorKind(t, err, tt.wantErr)
		})
	}
}

func TestRefresh(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockUserRepository)
		wantErr error
	}{
		{
			name: "success",
			setup: func(users *MockUserRepository) {
				user := sampleUser()
				users.EXPECT().RotateRefreshSession(gomock.Any(), "old-refresh", gomock.Any(), gomock.Any()).Return(user, nil)
			},
		},
		{
			name: "invalid token",
			setup: func(users *MockUserRepository) {
				users.EXPECT().RotateRefreshSession(gomock.Any(), "old-refresh", gomock.Any(), gomock.Any()).Return(store.User{}, store.ErrNotFound)
			},
			wantErr: ErrUnauthenticated,
		},
		{
			name: "rotation failure",
			setup: func(users *MockUserRepository) {
				users.EXPECT().RotateRefreshSession(gomock.Any(), "old-refresh", gomock.Any(), gomock.Any()).Return(store.User{}, errors.New("db down"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			users := NewMockUserRepository(ctrl)
			tt.setup(users)

			svc := New(testConfig(), users, NewMockVerificationStore(ctrl), NewMockEventPublisher(ctrl))
			_, err := svc.Refresh(context.Background(), dto.RefreshInput{RefreshToken: "old-refresh"})
			assertErrorKind(t, err, tt.wantErr)
		})
	}
}

func TestResendVerificationCode(t *testing.T) {
	verifiedAt := time.Now()
	tests := []struct {
		name    string
		setup   func(*MockUserRepository, *MockVerificationStore, *MockEventPublisher)
		wantErr error
	}{
		{
			name: "success",
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore, publisher *MockEventPublisher) {
				user := sampleUser()
				users.EXPECT().GetUserByID(gomock.Any(), user.ID).Return(user, nil)
				verificationStore.EXPECT().ResendCode(gomock.Any(), user.ID).Return(verification.CodeState{
					Code:              "123456",
					CodeExpiresAt:     time.Now().Add(time.Minute),
					ResendAvailableAt: time.Now().Add(time.Minute),
				}, nil)
				publisher.EXPECT().EmailVerificationRequested(gomock.Any(), user.ID, user.Email, "123456").Return(nil)
			},
		},
		{
			name: "cooldown",
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore, publisher *MockEventPublisher) {
				user := sampleUser()
				users.EXPECT().GetUserByID(gomock.Any(), user.ID).Return(user, nil)
				verificationStore.EXPECT().ResendCode(gomock.Any(), user.ID).Return(verification.CodeState{}, verification.ErrCooldown)
			},
			wantErr: ErrResourceExhausted,
		},
		{
			name: "lockout",
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore, publisher *MockEventPublisher) {
				user := sampleUser()
				users.EXPECT().GetUserByID(gomock.Any(), user.ID).Return(user, nil)
				verificationStore.EXPECT().ResendCode(gomock.Any(), user.ID).Return(verification.CodeState{}, verification.ErrLockout)
			},
			wantErr: ErrResourceExhausted,
		},
		{
			name: "already verified",
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore, publisher *MockEventPublisher) {
				user := sampleUser()
				user.EmailVerifiedAt = &verifiedAt
				users.EXPECT().GetUserByID(gomock.Any(), user.ID).Return(user, nil)
			},
			wantErr: ErrFailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			users := NewMockUserRepository(ctrl)
			verificationStore := NewMockVerificationStore(ctrl)
			publisher := NewMockEventPublisher(ctrl)
			tt.setup(users, verificationStore, publisher)

			svc := New(testConfig(), users, verificationStore, publisher)
			_, err := svc.ResendVerificationCode(context.Background(), dto.ResendVerificationCodeInput{UserID: sampleUser().ID})
			assertErrorKind(t, err, tt.wantErr)
		})
	}
}

func TestVerifyEmail(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockUserRepository, *MockVerificationStore)
		wantErr error
	}{
		{
			name: "success",
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore) {
				user := sampleUser()
				verificationStore.EXPECT().VerifyCode(gomock.Any(), user.ID, "123456").Return(nil)
				users.EXPECT().MarkEmailVerified(gomock.Any(), user.ID).Return(user, nil)
			},
		},
		{
			name: "bad code",
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore) {
				verificationStore.EXPECT().VerifyCode(gomock.Any(), sampleUser().ID, "bad").Return(verification.ErrBadCode)
			},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "lockout",
			setup: func(users *MockUserRepository, verificationStore *MockVerificationStore) {
				verificationStore.EXPECT().VerifyCode(gomock.Any(), sampleUser().ID, "123456").Return(verification.ErrLockout)
			},
			wantErr: ErrResourceExhausted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			users := NewMockUserRepository(ctrl)
			verificationStore := NewMockVerificationStore(ctrl)
			tt.setup(users, verificationStore)

			code := "123456"
			if tt.name == "bad code" {
				code = "bad"
			}
			svc := New(testConfig(), users, verificationStore, NewMockEventPublisher(ctrl))
			_, err := svc.VerifyEmail(context.Background(), dto.VerifyEmailInput{UserID: sampleUser().ID, Code: code})
			assertErrorKind(t, err, tt.wantErr)
		})
	}
}

func TestValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := NewMockUserRepository(ctrl)
	svc := New(testConfig(), users, NewMockVerificationStore(ctrl), NewMockEventPublisher(ctrl))

	token, _, err := svc.issueAccessToken(sampleUser())
	if err != nil {
		t.Fatal(err)
	}

	inactive := sampleUser()
	inactive.Status = "blocked"
	gomock.InOrder(
		users.EXPECT().GetUserByID(gomock.Any(), sampleUser().ID).Return(sampleUser(), nil),
		users.EXPECT().GetUserByID(gomock.Any(), sampleUser().ID).Return(inactive, nil),
	)

	tests := []struct {
		name      string
		token     string
		wantValid bool
	}{
		{name: "valid", token: token, wantValid: true},
		{name: "invalid", token: "not-a-token", wantValid: false},
		{name: "inactive user", token: token, wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := svc.ValidateToken(context.Background(), dto.ValidateTokenInput{AccessToken: tt.token})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.Valid != tt.wantValid {
				t.Fatalf("valid = %v, want %v", out.Valid, tt.wantValid)
			}
		})
	}
}

func TestBlockUser(t *testing.T) {
	tests := []struct {
		name    string
		input   dto.BlockUserInput
		setup   func(*MockUserRepository)
		wantErr error
	}{
		{
			name: "success",
			input: dto.BlockUserInput{
				ActorUserID:  99,
				ActorRole:    commonv1.UserRole_USER_ROLE_MODERATOR,
				UserID:       7,
				BlockedUntil: time.Now().Add(time.Hour),
			},
			setup: func(users *MockUserRepository) {
				user := sampleUser()
				blocked := user
				blocked.Status = "blocked"
				users.EXPECT().GetUserByID(gomock.Any(), int64(7)).Return(user, nil)
				users.EXPECT().BlockUser(gomock.Any(), int64(7), gomock.Any()).Return(blocked, nil)
			},
		},
		{
			name: "owner cannot block",
			input: dto.BlockUserInput{
				ActorUserID:  99,
				ActorRole:    commonv1.UserRole_USER_ROLE_OWNER,
				UserID:       7,
				BlockedUntil: time.Now().Add(time.Hour),
			},
			setup:   func(users *MockUserRepository) {},
			wantErr: ErrPermissionDenied,
		},
		{
			name: "cannot block self",
			input: dto.BlockUserInput{
				ActorUserID:  7,
				ActorRole:    commonv1.UserRole_USER_ROLE_MODERATOR,
				UserID:       7,
				BlockedUntil: time.Now().Add(time.Hour),
			},
			setup:   func(users *MockUserRepository) {},
			wantErr: ErrFailedPrecondition,
		},
		{
			name: "duration required",
			input: dto.BlockUserInput{
				ActorUserID:  99,
				ActorRole:    commonv1.UserRole_USER_ROLE_MODERATOR,
				UserID:       7,
				BlockedUntil: time.Now().Add(-time.Hour),
			},
			setup:   func(users *MockUserRepository) {},
			wantErr: ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			users := NewMockUserRepository(ctrl)
			tt.setup(users)

			svc := New(testConfig(), users, NewMockVerificationStore(ctrl), NewMockEventPublisher(ctrl))
			_, err := svc.BlockUser(context.Background(), tt.input)
			assertErrorKind(t, err, tt.wantErr)
		})
	}
}

func testConfig() config.Config {
	return config.Config{
		JWTSecret:  "test-secret",
		AccessTTL:  time.Hour,
		RefreshTTL: time.Hour,
	}
}

func sampleUser() store.User {
	now := time.Now()
	return store.User{
		ID:           7,
		Email:        "worker@example.com",
		PasswordHash: "hash",
		Role:         commonv1.UserRole_USER_ROLE_WORKER,
		LastName:     "User",
		FirstName:    "Worker",
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func assertErrorKind(t *testing.T, err error, want error) {
	t.Helper()
	if want == nil {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected error %v", want)
	}
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want kind %v", err, want)
	}
}
