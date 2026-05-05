package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/berusmenu/api-gateway/internal/config"
	"gitlab.com/berusmenu/api-gateway/internal/http/docs"
	authhandler "gitlab.com/berusmenu/api-gateway/internal/http/handlers/auth"
	communicationhandler "gitlab.com/berusmenu/api-gateway/internal/http/handlers/communication"
	inventoryhandler "gitlab.com/berusmenu/api-gateway/internal/http/handlers/inventory"
	matchinghandler "gitlab.com/berusmenu/api-gateway/internal/http/handlers/matching"
	"gitlab.com/berusmenu/api-gateway/internal/http/middleware"
	authv1 "gitlab.com/berusmenu/proto/gen/go/auth/v1"
	communicationv1 "gitlab.com/berusmenu/proto/gen/go/communication/v1"
	inventoryv1 "gitlab.com/berusmenu/proto/gen/go/inventory/v1"
	matchingv1 "gitlab.com/berusmenu/proto/gen/go/matching/v1"
)

type Dependencies struct {
	Config        config.Config
	Auth          authv1.AuthServiceClient
	Inventory     inventoryv1.InventoryServiceClient
	Matching      matchingv1.MatchingServiceClient
	Communication communicationv1.CommunicationServiceClient
}

func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()

	router.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.Timeout(deps.Config.RequestTimeout),
	)

	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/openapi.json", func(c *gin.Context) {
		spec, err := docs.OpenAPIJSON()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "internal",
					"message": "failed to build OpenAPI document",
				},
			})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", spec)
	})

	router.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(docs.SwaggerUIHTML))
	})

	authMiddleware := middleware.Auth(deps.Auth)
	verifiedMiddleware := middleware.RequireVerified()

	api := router.Group("/v1")
	authhandler.RegisterRoutes(api, deps.Auth, authMiddleware, verifiedMiddleware)
	inventoryhandler.RegisterRoutes(api, deps.Inventory, authMiddleware, verifiedMiddleware)
	matchinghandler.RegisterRoutes(api, deps.Matching, authMiddleware, verifiedMiddleware)
	communicationhandler.RegisterRoutes(router, api, deps.Communication, authMiddleware, verifiedMiddleware)

	return router
}
