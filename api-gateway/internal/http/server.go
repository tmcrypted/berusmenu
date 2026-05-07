package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/config"
	authhandler "github.com/tmcrypted/berusmenu/api-gateway/internal/http/handlers/auth"
	communicationhandler "github.com/tmcrypted/berusmenu/api-gateway/internal/http/handlers/communication"
	inventoryhandler "github.com/tmcrypted/berusmenu/api-gateway/internal/http/handlers/inventory"
	matchinghandler "github.com/tmcrypted/berusmenu/api-gateway/internal/http/handlers/matching"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/middleware"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
	communicationv1 "github.com/tmcrypted/berusmenu/proto/gen/go/communication/v1"
	inventoryv1 "github.com/tmcrypted/berusmenu/proto/gen/go/inventory/v1"
	matchingv1 "github.com/tmcrypted/berusmenu/proto/gen/go/matching/v1"
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
		middleware.SecurityHeaders(),
		middleware.CORS(deps.Config),
		middleware.Timeout(deps.Config.RequestTimeout),
	)

	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authMiddleware := middleware.Auth(deps.Auth, deps.Config)
	csrfMiddleware := middleware.CSRF(deps.Config)
	verifiedMiddleware := middleware.RequireVerified()

	api := router.Group("/v1")
	authhandler.RegisterRoutes(api, deps.Config, deps.Auth, authMiddleware, csrfMiddleware, verifiedMiddleware)
	inventoryhandler.RegisterRoutes(api, deps.Inventory, authMiddleware, csrfMiddleware, verifiedMiddleware)
	matchinghandler.RegisterRoutes(api, deps.Matching, authMiddleware, csrfMiddleware, verifiedMiddleware)
	communicationhandler.RegisterRoutes(router, api, deps.Communication, authMiddleware, csrfMiddleware, verifiedMiddleware)

	return router
}
