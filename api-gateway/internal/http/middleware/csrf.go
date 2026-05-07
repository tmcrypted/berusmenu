package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/config"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/response"
)

func CSRF(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requiresCSRF(c.Request.Method) {
			c.Next()
			return
		}

		cookieToken, err := c.Cookie(cfg.CSRFCookieName)
		if err != nil || cookieToken == "" {
			response.Error(c, http.StatusForbidden, "csrf_required", "missing csrf cookie")
			return
		}

		headerToken := c.GetHeader(cfg.CSRFHeaderName)
		if headerToken == "" || headerToken != cookieToken {
			response.Error(c, http.StatusForbidden, "csrf_invalid", "invalid csrf token")
			return
		}

		c.Next()
	}
}

func requiresCSRF(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}
