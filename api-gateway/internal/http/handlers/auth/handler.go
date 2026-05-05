package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.com/berusmenu/api-gateway/internal/http/middleware"
	"gitlab.com/berusmenu/api-gateway/internal/http/response"
	authv1 "gitlab.com/berusmenu/proto/gen/go/auth/v1"
	commonv1 "gitlab.com/berusmenu/proto/gen/go/common/v1"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	auth authv1.AuthServiceClient
}

func RegisterRoutes(
	group *gin.RouterGroup,
	client authv1.AuthServiceClient,
	authMiddleware gin.HandlerFunc,
	verifiedMiddleware gin.HandlerFunc,
) {
	handler := Handler{auth: client}

	group.POST("/auth/register", handler.Register)
	group.POST("/auth/login", handler.Login)
	group.POST("/auth/refresh", handler.Refresh)
	group.POST("/auth/logout", authMiddleware, handler.Logout)
	group.POST("/auth/resend-code", authMiddleware, handler.ResendVerificationCode)
	group.POST("/auth/verify-email", authMiddleware, handler.VerifyEmail)

	group.GET("/me", authMiddleware, verifiedMiddleware, handler.GetMe)
	group.PATCH("/me", authMiddleware, verifiedMiddleware, handler.UpdateMe)
}

type registerRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=8"`
	FullName string  `json:"full_name" binding:"required"`
	Role     string  `json:"role"`
	Phone    *string `json:"phone"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type updateProfileRequest struct {
	FullName  *string `json:"full_name"`
	Phone     *string `json:"phone"`
	AvatarURL *string `json:"avatar_url"`
}

type verifyEmailRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h Handler) Register(c *gin.Context) {
	var req registerRequest
	if !bindJSON(c, &req) {
		return
	}

	role, err := ParseUserRole(req.Role)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_argument", err.Error())
		return
	}

	res, err := h.auth.Register(c.Request.Context(), &authv1.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     role,
		Phone:    req.Phone,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}

	writeAuthPayload(c, http.StatusCreated, res.GetAuth())
}

func (h Handler) Login(c *gin.Context) {
	var req loginRequest
	if !bindJSON(c, &req) {
		return
	}

	res, err := h.auth.Login(c.Request.Context(), &authv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}

	writeAuthPayload(c, http.StatusOK, res.GetAuth())
}

func (h Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if !bindJSON(c, &req) {
		return
	}

	res, err := h.auth.Refresh(c.Request.Context(), &authv1.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}

	writeAuthPayload(c, http.StatusOK, res.GetAuth())
}

func (h Handler) Logout(c *gin.Context) {
	var req logoutRequest
	if !bindJSON(c, &req) {
		return
	}

	if _, err := h.auth.Logout(c.Request.Context(), &authv1.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		response.FromGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h Handler) ResendVerificationCode(c *gin.Context) {
	res, err := h.auth.ResendVerificationCode(c.Request.Context(), &authv1.ResendVerificationCodeRequest{})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}

	writeProto(c, http.StatusOK, res)
}

func (h Handler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if !bindJSON(c, &req) {
		return
	}

	res, err := h.auth.VerifyEmail(c.Request.Context(), &authv1.VerifyEmailRequest{
		Code: req.Code,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}

	writeProto(c, http.StatusOK, res)
}

func (h Handler) GetMe(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthenticated", "missing authenticated user")
		return
	}

	res, err := h.auth.GetUser(c.Request.Context(), &authv1.GetUserRequest{
		UserId: user.ID,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}

	writeProto(c, http.StatusOK, res.GetUser())
}

func (h Handler) UpdateMe(c *gin.Context) {
	var req updateProfileRequest
	if !bindJSON(c, &req) {
		return
	}

	res, err := h.auth.UpdateProfile(c.Request.Context(), &authv1.UpdateProfileRequest{
		FullName:  req.FullName,
		Phone:     req.Phone,
		AvatarUrl: req.AvatarURL,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}

	writeProto(c, http.StatusOK, res.GetUser())
}

func ParseUserRole(value string) (commonv1.UserRole, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "worker":
		return commonv1.UserRole_USER_ROLE_WORKER, nil
	case "owner":
		return commonv1.UserRole_USER_ROLE_OWNER, nil
	case "moderator":
		return commonv1.UserRole_USER_ROLE_MODERATOR, nil
	default:
		return commonv1.UserRole_USER_ROLE_UNSPECIFIED, errors.New("role must be one of: owner, worker, moderator")
	}
}

func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_argument", err.Error())
		return false
	}
	return true
}

func writeAuthPayload(c *gin.Context, status int, payload *authv1.AuthPayload) {
	writeProto(c, status, payload)
}

func writeProto(c *gin.Context, status int, msg proto.Message) {
	if msg == nil {
		response.Error(c, http.StatusInternalServerError, "internal", "empty upstream response")
		return
	}
	response.Proto(c, status, msg)
}
