package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		requestID := GetRequestID(c.Request.Context())
		c.Next()
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()
		slog.Info("http request",
			"request_id", requestID,
			"method", method,
			"path", path,
			"status", statusCode,
			"latency_ms", latency.Milliseconds(),
			"client_ip", clientIP,
			"user_agent", userAgent,
		)
	}
}