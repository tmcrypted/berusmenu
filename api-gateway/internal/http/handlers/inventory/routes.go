package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/berusmenu/api-gateway/internal/http/response"
	inventoryv1 "gitlab.com/berusmenu/proto/gen/go/inventory/v1"
)

func RegisterRoutes(
	group *gin.RouterGroup,
	client inventoryv1.InventoryServiceClient,
	authMiddleware gin.HandlerFunc,
	verifiedMiddleware gin.HandlerFunc,
) {
	_ = client

	authorized := group.Group("", authMiddleware, verifiedMiddleware)
	authorized.POST("/pvz", notImplemented())
	authorized.GET("/pvz", notImplemented())
	authorized.GET("/pvz/:id", notImplemented())
	authorized.PATCH("/pvz/:id", notImplemented())
	authorized.POST("/pvz/:id/submit-update", notImplemented())

	moderation := authorized.Group("/moderation")
	moderation.GET("/pvz", notImplemented())
	moderation.POST("/pvz/:id/approve", notImplemented())
	moderation.POST("/pvz/:id/reject", notImplemented())
	moderation.POST("/pvz/:id/block", notImplemented())
}

func notImplemented() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusNotImplemented, "not_implemented", "inventory gateway routes are not implemented yet")
	}
}
