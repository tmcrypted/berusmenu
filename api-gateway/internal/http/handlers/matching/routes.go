package matching

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/berusmenu/api-gateway/internal/http/response"
	matchingv1 "gitlab.com/berusmenu/proto/gen/go/matching/v1"
)

func RegisterRoutes(
	group *gin.RouterGroup,
	client matchingv1.MatchingServiceClient,
	authMiddleware gin.HandlerFunc,
	verifiedMiddleware gin.HandlerFunc,
) {
	_ = client

	authorized := group.Group("", authMiddleware, verifiedMiddleware)
	authorized.POST("/vacancies", notImplemented())
	authorized.GET("/vacancies", notImplemented())
	authorized.GET("/vacancies/:id", notImplemented())
	authorized.POST("/vacancies/:id/responses", notImplemented())
	authorized.GET("/vacancies/:id/responses", notImplemented())
	authorized.POST("/responses/:id/withdraw", notImplemented())
	authorized.POST("/responses/:id/accept", notImplemented())
	authorized.POST("/assignments/:id/cancel", notImplemented())
	authorized.POST("/assignments/:id/complete", notImplemented())
	authorized.POST("/assignments/:id/reviews", notImplemented())
	authorized.GET("/reviews", notImplemented())
	authorized.POST("/reviews/:id/replies", notImplemented())
}

func notImplemented() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusNotImplemented, "not_implemented", "matching gateway routes are not implemented yet")
	}
}
