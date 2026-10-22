package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/shop/backend/internal/transport/http/dto"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(c.Request.Context())
				slog.Error("panic recovered",
					"request_id", requestID,
					"error", err,
					"stack", string(debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
					Error: dto.ErrorBody{
						Code:    "internal",
						Message: "internal server error",
					},
				})
			}
		}()
		c.Next()
	}
}