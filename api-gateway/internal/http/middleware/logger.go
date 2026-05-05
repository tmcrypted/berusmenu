package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/berusmenu/api-gateway/internal/http/response"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		log.Printf(
			"request_id=%s method=%s path=%s status=%d latency=%s client_ip=%s",
			response.RequestID(c),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
			c.ClientIP(),
		)
	}
}
