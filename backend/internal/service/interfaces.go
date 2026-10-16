package service

import (
	"context"

	"github.com/shop/backend/internal/domain"
)

type ProductService interface {
	Create(ctx context.Context, product *domain.Product) (*domain.Product, error)
	GetAll(ctx context.Context, page, pageSize int) ([]domain.Product, error)
	GetByID(ctx context.Context, id uint) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) (*domain.Product, error)
	Delete(ctx context.Context, id uint) error
}

type OrderService interface {
	Create(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetAll(ctx context.Context, page, pageSize int) ([]domain.Order, error)
	GetByID(ctx context.Context, id uint) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) (*domain.Order, error)
	Delete(ctx context.Context, id uint) error
}