package middleware

import (
	"log"
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

		log.Printf("[GIN] request_id=%s %3d | %13v | %15s | %-7s %s | %s",
			requestID, statusCode, latency, clientIP, method, path, userAgent)
	}
}