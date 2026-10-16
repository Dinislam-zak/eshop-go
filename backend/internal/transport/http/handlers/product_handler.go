package handlers

import (
	"net/http"
	"strconv"

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

	product := &domain.Product{
		Name:  req.Name,
		Price: req.Price,
	}

	created, err := h.service.Create(c.Request.Context(), product)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toProductResponse(created))
}

func (h *ProductHandler) GetAll(c *gin.Context) {
	products, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	response := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		response[i] = toProductResponse(&p)
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	product, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toProductResponse(product))
}

func (h *ProductHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindingError(c, err)
		return
	}

	product := &domain.Product{
		ID:    uint(id),
		Name:  req.Name,
		Price: req.Price,
	}

	updated, err := h.service.Update(c.Request.Context(), product)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toProductResponse(updated))
}

func (h *ProductHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
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