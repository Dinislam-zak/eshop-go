package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateProductRequest struct {
	Name  string          `json:"name" binding:"required"`
	Price decimal.Decimal `json:"price" binding:"required"`
}

type UpdateProductRequest struct {
	Name  string          `json:"name" binding:"required"`
	Price decimal.Decimal `json:"price" binding:"required"`
}

type ProductResponse struct {
	ID        uint            `json:"id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}