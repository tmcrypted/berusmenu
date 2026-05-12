package dto

type RegisterRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=8"`
	Role     string  `json:"role"`
	Phone    *string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateProfileRequest struct {
	LastName         *string `json:"last_name"`
	FirstName        *string `json:"first_name"`
	MiddleName       *string `json:"middle_name"`
	BirthDate        *string `json:"birth_date"`
	Phone            *string `json:"phone"`
	AvatarURL        *string `json:"avatar_url"`
	TelegramUsername *string `json:"telegram_username"`
	MaxUsername      *string `json:"max_username"`
}

type VerifyEmailRequest struct {
	Code string `json:"code" binding:"required"`
}
