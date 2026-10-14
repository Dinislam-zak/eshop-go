package dto

import "time"

type OrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	Number          string             `json:"number" binding:"required"`
	DeliveryAddress string             `json:"delivery_address" binding:"required"`
	Items           []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type UpdateOrderRequest struct {
	Number          string             `json:"number" binding:"required"`
	DeliveryAddress string             `json:"delivery_address" binding:"required"`
	Items           []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type OrderItemResponse struct {
	ProductID uint            `json:"product_id"`
	Product   ProductResponse `json:"product"`
	Quantity  int             `json:"quantity"`
}

type OrderResponse struct {
	ID              uint                `json:"id"`
	Number          string              `json:"number"`
	DeliveryAddress string              `json:"delivery_address"`
	Items           []OrderItemResponse `json:"items"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}