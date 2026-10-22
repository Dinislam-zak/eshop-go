package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shop/backend/internal/domain"
	"github.com/shop/backend/internal/service"
	"github.com/shop/backend/internal/transport/http/dto"
)

type OrderHandler struct {
	service service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindingError(c, err)
		return
	}
	order := &domain.Order{
		Number:          req.Number,
		DeliveryAddress: req.DeliveryAddress,
	}
	for _, item := range req.Items {
		order.Items = append(order.Items, domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	created, err := h.service.Create(c.Request.Context(), order)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toOrderResponse(created))
}

func (h *OrderHandler) GetAll(c *gin.Context) {
	var req dto.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		handleBindingError(c, err)
		return
	}
	req.Normalize()
	orders, err := h.service.GetAll(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	response := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		response[i] = toOrderResponse(&o)
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse[dto.OrderResponse]{
		Items:    response,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	order, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOrderResponse(order))
}

func (h *OrderHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindingError(c, err)
		return
	}
	order := &domain.Order{
		ID:              id,
		Number:          req.Number,
		DeliveryAddress: req.DeliveryAddress,
	}
	for _, item := range req.Items {
		order.Items = append(order.Items, domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	updated, err := h.service.Update(c.Request.Context(), order)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOrderResponse(updated))
}

func (h *OrderHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func toOrderResponse(o *domain.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = dto.OrderItemResponse{
			ProductID: item.ProductID,
			Product:   toProductResponse(&item.Product),
			Quantity:  item.Quantity,
		}
	}
	return dto.OrderResponse{
		ID:              o.ID,
		Number:          o.Number,
		DeliveryAddress: o.DeliveryAddress,
		Items:           items,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
}