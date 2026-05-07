package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/response"
	inventoryv1 "github.com/tmcrypted/berusmenu/proto/gen/go/inventory/v1"
)

func RegisterRoutes(
	group *gin.RouterGroup,
	client inventoryv1.InventoryServiceClient,
	authMiddleware gin.HandlerFunc,
	csrfMiddleware gin.HandlerFunc,
	verifiedMiddleware gin.HandlerFunc,
) {
	_ = client

	authorized := group.Group("", authMiddleware, verifiedMiddleware)
	authorized.POST("/pvz", csrfMiddleware, notImplemented())
	authorized.GET("/pvz", notImplemented())
	authorized.GET("/pvz/:id", notImplemented())
	authorized.PATCH("/pvz/:id", csrfMiddleware, notImplemented())

	moderation := authorized.Group("/moderation")
	moderation.GET("/pvz", notImplemented())
	moderation.POST("/pvz/:id/approve", csrfMiddleware, notImplemented())
	moderation.POST("/pvz/:id/reject", csrfMiddleware, notImplemented())
	moderation.POST("/pvz/:id/block", csrfMiddleware, notImplemented())
}

func notImplemented() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusNotImplemented, "not_implemented", "inventory gateway routes are not implemented yet")
	}
}
