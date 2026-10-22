package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		
		ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}