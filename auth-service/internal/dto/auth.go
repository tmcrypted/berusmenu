package dto

import (
	"time"

	"github.com/tmcrypted/berusmenu/auth-service/internal/store"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
)

type RegisterInput struct {
	Email    string
	Password string
	FullName string
	Role     commonv1.UserRole
	Phone    *string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshInput struct {
	RefreshToken string
}

type LogoutInput struct {
	RefreshToken string
}

type ValidateTokenInput struct {
	AccessToken string
}

type GetUserInput struct {
	CurrentUserID int64
	UserID        int64
}

type UpdateProfileInput struct {
	UserID    int64
	FullName  *string
	Phone     *string
	AvatarURL *string
}

type ResendVerificationCodeInput struct {
	UserID int64
}

type VerifyEmailInput struct {
	UserID int64
	Code   string
}

type AuthPayload struct {
	User            store.User
	AccessToken     string
	RefreshToken    string
	AccessExpiresAt time.Time
}

type ValidateTokenOutput struct {
	Valid         bool
	UserID        int64
	Role          commonv1.UserRole
	Email         string
	EmailVerified bool
}

type VerificationCodeOutput struct {
	CodeExpiresAt     time.Time
	ResendAvailableAt time.Time
}
