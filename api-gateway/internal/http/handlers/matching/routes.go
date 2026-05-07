package matching

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/response"
	matchingv1 "github.com/tmcrypted/berusmenu/proto/gen/go/matching/v1"
)

func RegisterRoutes(
	group *gin.RouterGroup,
	client matchingv1.MatchingServiceClient,
	authMiddleware gin.HandlerFunc,
	csrfMiddleware gin.HandlerFunc,
	verifiedMiddleware gin.HandlerFunc,
) {
	_ = client

	authorized := group.Group("", authMiddleware, verifiedMiddleware)
	authorized.POST("/vacancies", csrfMiddleware, notImplemented())
	authorized.GET("/vacancies", notImplemented())
	authorized.GET("/vacancies/:id", notImplemented())
	authorized.POST("/vacancies/:id/responses", csrfMiddleware, notImplemented())
	authorized.GET("/vacancies/:id/responses", notImplemented())
	authorized.POST("/responses/:id/withdraw", csrfMiddleware, notImplemented())
	authorized.POST("/responses/:id/accept", csrfMiddleware, notImplemented())
	authorized.POST("/assignments/:id/cancel", csrfMiddleware, notImplemented())
	authorized.POST("/assignments/:id/complete", csrfMiddleware, notImplemented())
	authorized.POST("/assignments/:id/reviews", csrfMiddleware, notImplemented())
	authorized.GET("/reviews", notImplemented())
	authorized.POST("/reviews/:id/replies", csrfMiddleware, notImplemented())
}

func notImplemented() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusNotImplemented, "not_implemented", "matching gateway routes are not implemented yet")
	}
}
