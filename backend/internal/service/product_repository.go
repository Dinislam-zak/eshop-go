package service

import (
	"context"

	"github.com/shop/backend/internal/domain"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetAll(ctx context.Context, page, pageSize int) ([]domain.Product, error)
	GetByID(ctx context.Context, id uint) (*domain.Product, error)
	GetByIDs(ctx context.Context, ids []uint) ([]domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id uint) error
}