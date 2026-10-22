package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shop/backend/internal/domain"
	"github.com/shop/backend/internal/transport/http/middleware"
)

func handleError(c *gin.Context, err error) {
	requestID := middleware.GetRequestID(c.Request.Context())

	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInUse):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		log.Printf("[ERROR] request_id=%s unhandled error: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func handleBindingError(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		response := make(map[string]string)
		for _, e := range validationErrors {
			response[e.Field()] = e.Tag()
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": response})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}