package service

import (
	"context"

	"github.com/shop/backend/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetAll(ctx context.Context, page, pageSize int) ([]domain.Order, error)
	GetByID(ctx context.Context, id uint) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, id uint) error
}