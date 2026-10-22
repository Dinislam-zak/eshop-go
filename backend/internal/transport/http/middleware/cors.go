package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(origins []string) gin.HandlerFunc {
	allowedOrigins := make(map[string]bool)
	allowAll := false
	for _, origin := range origins {
		if origin == "*" {
			allowAll = true
		} else {
			allowedOrigins[strings.ToLower(origin)] = true
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if c.Request.Method == http.MethodOptions {
			if origin != "" && (allowAll || allowedOrigins[strings.ToLower(origin)]) {
				setCorsHeaders(c, origin, allowAll)
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		if origin == "" {
			c.Next()
			return
		}

		originLower := strings.ToLower(origin)
		if !allowAll && !allowedOrigins[originLower] {
			c.Next()
			return
		}

		setCorsHeaders(c, origin, allowAll)
		c.Next()
	}
}

func setCorsHeaders(c *gin.Context, origin string, allowAll bool) {
	if allowAll {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", origin)
	}
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	c.Header("Access-Control-Expose-Headers", "X-Request-ID")
	c.Header("Access-Control-Max-Age", "86400")

	if !allowAll {
		c.Header("Access-Control-Allow-Credentials", "true")
	}
}