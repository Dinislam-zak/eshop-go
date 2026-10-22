package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shop/backend/internal/domain"
	"github.com/shop/backend/internal/service"
	"github.com/shop/backend/internal/transport/http/dto"
)

type ProductHandler struct {
	service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindingError(c, err)
		return
	}
	product := &domain.Product{Name: req.Name, Price: req.Price}
	created, err := h.service.Create(c.Request.Context(), product)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toProductResponse(created))
}

func (h *ProductHandler) GetAll(c *gin.Context) {
	var req dto.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		handleBindingError(c, err)
		return
	}
	req.Normalize()
	products, err := h.service.GetAll(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	response := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		response[i] = toProductResponse(&p)
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse[dto.ProductResponse]{
		Items:    response,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	product, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProductResponse(product))
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindingError(c, err)
		return
	}
	product := &domain.Product{ID: id, Name: req.Name, Price: req.Price}
	updated, err := h.service.Update(c.Request.Context(), product)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProductResponse(updated))
}

func (h *ProductHandler) Delete(c *gin.Context) {
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

func toProductResponse(p *domain.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}