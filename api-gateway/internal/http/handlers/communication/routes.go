package communication

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/berusmenu/api-gateway/internal/http/response"
	communicationv1 "gitlab.com/berusmenu/proto/gen/go/communication/v1"
)

func RegisterRoutes(
	router *gin.Engine,
	group *gin.RouterGroup,
	client communicationv1.CommunicationServiceClient,
	authMiddleware gin.HandlerFunc,
	verifiedMiddleware gin.HandlerFunc,
) {
	_ = client

	authorized := group.Group("", authMiddleware, verifiedMiddleware)
	authorized.GET("/conversations", notImplemented())
	authorized.GET("/conversations/:id/messages", notImplemented())
	authorized.POST("/conversations/:id/messages", notImplemented())
	authorized.POST("/conversations/:id/read", notImplemented())

	ws := router.Group("/ws", authMiddleware, verifiedMiddleware)
	ws.GET("/conversations/:id", notImplemented())
}

func notImplemented() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusNotImplemented, "not_implemented", "communication gateway routes are not implemented yet")
	}
}
