package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/config"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/dto"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/middleware"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/response"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	auth    authv1.AuthServiceClient
	cookies cookieManager
}

type authPayloadResponse struct {
	User            json.RawMessage `json:"user"`
	AccessExpiresAt string          `json:"access_expires_at"`
}

func RegisterRoutes(
	group *gin.RouterGroup,
	cfg config.Config,
	client authv1.AuthServiceClient,
	authMiddleware gin.HandlerFunc,
	csrfMiddleware gin.HandlerFunc,
	verifiedMiddleware gin.HandlerFunc,
) {
	handler := Handler{auth: client, cookies: newCookieManager(cfg)}

	group.POST("/auth/register", handler.Register)
	group.POST("/auth/login", handler.Login)
	group.POST("/auth/refresh", handler.Refresh)
	group.GET("/auth/session", authMiddleware, handler.GetSession)
	group.POST("/auth/logout", csrfMiddleware, handler.Logout)
	group.POST("/auth/resend-code", authMiddleware, csrfMiddleware, handler.ResendVerificationCode)
	group.POST("/auth/verify-email", authMiddleware, csrfMiddleware, handler.VerifyEmail)

	group.GET("/me", authMiddleware, verifiedMiddleware, handler.GetMe)
	group.PATCH("/me", authMiddleware, verifiedMiddleware, csrfMiddleware, handler.UpdateMe)
}

func (h Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
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

	h.writeAuthPayload(c, http.StatusCreated, res.GetAuth())
}

func (h Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
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

	h.writeAuthPayload(c, http.StatusOK, res.GetAuth())
}

func (h Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(h.cookies.cfg.RefreshCookieName)
	if err != nil || refreshToken == "" {
		response.Error(c, http.StatusUnauthorized, "unauthenticated", "missing refresh cookie")
		return
	}

	res, err := h.auth.Refresh(c.Request.Context(), &authv1.RefreshRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}

	h.writeAuthPayload(c, http.StatusOK, res.GetAuth())
}

func (h Handler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie(h.cookies.cfg.RefreshCookieName)
	if err != nil || refreshToken == "" {
		response.Error(c, http.StatusUnauthorized, "unauthenticated", "missing refresh cookie")
		return
	}

	if _, err := h.auth.Logout(c.Request.Context(), &authv1.LogoutRequest{
		RefreshToken: refreshToken,
	}); err != nil {
		response.FromGRPCError(c, err)
		return
	}

	h.cookies.clearAuthCookies(c)
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
	var req dto.VerifyEmailRequest
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

func (h Handler) GetSession(c *gin.Context) {
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
	var req dto.UpdateProfileRequest
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

func (h Handler) writeAuthPayload(c *gin.Context, status int, payload *authv1.AuthPayload) {
	if payload == nil || payload.GetTokens() == nil {
		response.Error(c, http.StatusInternalServerError, "internal", "empty upstream auth response")
		return
	}
	if err := h.cookies.setAuthCookies(c, payload.GetTokens()); err != nil {
		response.Error(c, http.StatusInternalServerError, "internal", "failed to create auth cookies")
		return
	}

	userJSON, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(payload.GetUser())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal", "failed to encode auth response")
		return
	}
	c.JSON(status, authPayloadResponse{
		User:            userJSON,
		AccessExpiresAt: payload.GetTokens().GetAccessExpiresAt().AsTime().UTC().Format(time.RFC3339),
	})
}

func writeProto(c *gin.Context, status int, msg proto.Message) {
	if msg == nil {
		response.Error(c, http.StatusInternalServerError, "internal", "empty upstream response")
		return
	}
	response.Proto(c, status, msg)
}
