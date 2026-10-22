package http

import (
	"github.com/gin-gonic/gin"
	"github.com/shop/backend/internal/config"
	"github.com/shop/backend/internal/service"
	"github.com/shop/backend/internal/transport/http/handlers"
	"github.com/shop/backend/internal/transport/http/middleware"
)

func NewRouter(cfg *config.Config, productService service.ProductService, orderService service.OrderService) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.CORS(cfg.CORSOrigins), middleware.Logger(), middleware.Recovery())

	productHandler := handlers.NewProductHandler(productService)
	orderHandler := handlers.NewOrderHandler(orderService)

	api := r.Group("/api")
	{
		products := api.Group("/products")
		{
			products.POST("", productHandler.Create)
			products.GET("", productHandler.GetAll)
			products.GET("/:id", productHandler.GetByID)
			products.PUT("/:id", productHandler.Update)
			products.DELETE("/:id", productHandler.Delete)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.Create)
			orders.GET("", orderHandler.GetAll)
			orders.GET("/:id", orderHandler.GetByID)
			orders.PUT("/:id", orderHandler.Update)
			orders.DELETE("/:id", orderHandler.Delete)
		}
	}

	return r
}