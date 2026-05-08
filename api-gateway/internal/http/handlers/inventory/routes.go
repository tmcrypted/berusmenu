package inventory

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/dto"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/response"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
	inventoryv1 "github.com/tmcrypted/berusmenu/proto/gen/go/inventory/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	inventory inventoryv1.InventoryServiceClient
	auth      authv1.AuthServiceClient
}

func RegisterRoutes(
	group *gin.RouterGroup,
	inventoryClient inventoryv1.InventoryServiceClient,
	authClient authv1.AuthServiceClient,
	authMiddleware gin.HandlerFunc,
	csrfMiddleware gin.HandlerFunc,
	verifiedMiddleware gin.HandlerFunc,
) {
	handler := Handler{inventory: inventoryClient, auth: authClient}

	authorized := group.Group("", authMiddleware, verifiedMiddleware)
	authorized.POST("/pvz", csrfMiddleware, handler.CreatePvz)
	authorized.GET("/pvz", handler.ListOwnerPvz)
	authorized.GET("/pvz/:id", handler.GetPvz)
	authorized.PATCH("/pvz/:id", csrfMiddleware, handler.UpdatePvz)

	moderation := authorized.Group("/moderation")
	moderation.GET("/pvz", handler.ListPendingPvz)
	moderation.POST("/pvz/:id/approve", csrfMiddleware, handler.ApprovePvz)
	moderation.POST("/pvz/:id/reject", csrfMiddleware, handler.RejectPvz)
	moderation.POST("/pvz/:id/block", csrfMiddleware, handler.BlockPvz)
	moderation.POST("/users/:id/block-and-delete-pvz", csrfMiddleware, handler.BlockUserAndDeletePvz)
}

func (h Handler) CreatePvz(c *gin.Context) {
	var req dto.CreatePvzRequest
	if !bindJSON(c, &req) {
		return
	}

	res, err := h.inventory.CreatePvz(c.Request.Context(), &inventoryv1.CreatePvzRequest{
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
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusCreated, res.GetPvz())
}

func (h Handler) ListOwnerPvz(c *gin.Context) {
	pageSize, ok := pageSize(c)
	if !ok {
		return
	}
	req := &inventoryv1.ListOwnerPvzRequest{PageSize: pageSize}
	if token := c.Query("page_token"); token != "" {
		req.PageToken = &token
	}

	res, err := h.inventory.ListOwnerPvz(c.Request.Context(), req)
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res)
}

func (h Handler) GetPvz(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	res, err := h.inventory.GetPvz(c.Request.Context(), &inventoryv1.GetPvzRequest{Id: id})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetPvz())
}

func (h Handler) UpdatePvz(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdatePvzRequest
	if !bindJSON(c, &req) {
		return
	}

	res, err := h.inventory.UpdatePvz(c.Request.Context(), &inventoryv1.UpdatePvzRequest{
		Id:              id,
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
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetPvz())
}

func (h Handler) ListPendingPvz(c *gin.Context) {
	pageSize, ok := pageSize(c)
	if !ok {
		return
	}
	req := &inventoryv1.ListPendingPvzRequest{PageSize: pageSize}
	if token := c.Query("page_token"); token != "" {
		req.PageToken = &token
	}

	res, err := h.inventory.ListPendingPvz(c.Request.Context(), req)
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res)
}

func (h Handler) ApprovePvz(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.ModerationCommentRequest
	if !bindOptionalJSON(c, &req) {
		return
	}
	res, err := h.inventory.ApprovePvz(c.Request.Context(), &inventoryv1.ApprovePvzRequest{PvzId: id, Comment: req.Comment})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetPvz())
}

func (h Handler) RejectPvz(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.RejectPvzRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.inventory.RejectPvz(c.Request.Context(), &inventoryv1.RejectPvzRequest{
		PvzId:   id,
		Reason:  req.Reason,
		Comment: req.Comment,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetPvz())
}

func (h Handler) BlockPvz(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.ModerationCommentRequest
	if !bindOptionalJSON(c, &req) {
		return
	}
	res, err := h.inventory.BlockPvz(c.Request.Context(), &inventoryv1.BlockPvzRequest{PvzId: id, Comment: req.Comment})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetPvz())
}

func (h Handler) BlockUserAndDeletePvz(c *gin.Context) {
	ownerID, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.BlockUserAndDeletePvzRequest
	if !bindJSON(c, &req) {
		return
	}
	blockedUntil := time.Now().Add(time.Duration(req.BlockDurationMinutes) * time.Minute)

	if _, err := h.auth.BlockUser(c.Request.Context(), &authv1.BlockUserRequest{
		UserId:       ownerID,
		Reason:       &req.Reason,
		BlockedUntil: timestamppb.New(blockedUntil),
	}); err != nil {
		response.FromGRPCError(c, err)
		return
	}

	res, err := h.inventory.DeleteOwnerPvzRequests(c.Request.Context(), &inventoryv1.DeleteOwnerPvzRequestsRequest{
		OwnerId: ownerID,
		Comment: req.Comment,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blocked_user_id": strconv.FormatInt(ownerID, 10),
		"blocked_until":   blockedUntil.UTC().Format(time.RFC3339),
		"deleted_count":   strconv.FormatInt(res.GetDeletedCount(), 10),
		"blocked_count":   strconv.FormatInt(res.GetBlockedCount(), 10),
	})
}

func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_argument", err.Error())
		return false
	}
	return true
}

func bindOptionalJSON(c *gin.Context, dst any) bool {
	if c.Request.Body == nil || c.Request.ContentLength == 0 {
		return true
	}
	return bindJSON(c, dst)
}

func pathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, "invalid_argument", "invalid "+name)
		return 0, false
	}
	return id, true
}

func pageSize(c *gin.Context) (int32, bool) {
	raw := c.Query("page_size")
	if raw == "" {
		return 0, true
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || value < 0 {
		response.Error(c, http.StatusBadRequest, "invalid_argument", "invalid page_size")
		return 0, false
	}
	return int32(value), true
}
