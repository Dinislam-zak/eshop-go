package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shop/backend/internal/domain"
	"github.com/shop/backend/internal/transport/http/dto"
	"github.com/shop/backend/internal/transport/http/middleware"
)

func handleError(c *gin.Context, err error) {
	requestID := middleware.GetRequestID(c.Request.Context())
	resp := dto.ErrorResponse{}
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		field := validationErr.Field
		resp.Error = dto.ErrorBody{
			Code:    "validation",
			Message: "validation failed",
			Details: []dto.ErrorDetail{{Field: field, Message: validationErr.Message}},
		}
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		resp.Error = dto.ErrorBody{Code: "not_found", Message: "entity not found"}
		c.JSON(http.StatusNotFound, resp)
	case errors.Is(err, domain.ErrAlreadyExists):
		resp.Error = dto.ErrorBody{Code: "already_exists", Message: "entity already exists"}
		c.JSON(http.StatusConflict, resp)
	case errors.Is(err, domain.ErrInUse):
		resp.Error = dto.ErrorBody{Code: "in_use", Message: "entity is in use"}
		c.JSON(http.StatusConflict, resp)
	default:
		slog.Error("unhandled error", "request_id", requestID, "error", err)
		resp.Error = dto.ErrorBody{Code: "internal", Message: "internal server error"}
		c.JSON(http.StatusInternalServerError, resp)
	}
}

func handleBindingError(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		resp := dto.ErrorResponse{}
		resp.Error.Code = "validation"
		resp.Error.Message = "validation failed"
		details := make([]dto.ErrorDetail, 0, len(validationErrors))
		for _, e := range validationErrors {
			details = append(details, dto.ErrorDetail{
				Field:   e.Field(),
				Message: e.Tag(),
			})
		}
		resp.Error.Details = details
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	resp := dto.ErrorResponse{}
	resp.Error = dto.ErrorBody{Code: "validation", Message: "invalid request payload"}
	c.JSON(http.StatusBadRequest, resp)
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	value, ok := c.Params.Get(name)
	if !ok || value == "" {
		field := name
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorBody{
				Code:    "validation",
				Message: "invalid " + name,
				Field:   &field,
			},
		})
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		field := name
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorBody{
				Code:    "validation",
				Message: "invalid " + name,
				Field:   &field,
			},
		})
		return 0, false
	}
	return uint(parsed), true
}