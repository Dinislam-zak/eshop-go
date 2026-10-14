package service

import (
	"context"
	"fmt"

	"github.com/shop/backend/internal/domain"
	"github.com/shop/backend/internal/repository"
)

type orderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
}

func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s *orderService) Create(ctx context.Context, order *domain.Order) error {
	if order.Number == "" {
		return &domain.ValidationError{Field: "number", Message: "number is required"}
	}
	if order.DeliveryAddress == "" {
		return &domain.ValidationError{Field: "delivery_address", Message: "delivery address is required"}
	}
	if len(order.Items) == 0 {
		return &domain.ValidationError{Field: "items", Message: "order must contain at least one item"}
	}

	ids := make([]uint, len(order.Items))
	seen := make(map[uint]bool)
	for i, item := range order.Items {
		if item.Quantity <= 0 {
			return &domain.ValidationError{Field: "items.quantity", Message: "quantity must be greater than zero"}
		}
		if seen[item.ProductID] {
			return &domain.ValidationError{Field: "items.product_id", Message: fmt.Sprintf("duplicate product id %d", item.ProductID)}
		}
		seen[item.ProductID] = true
		ids[i] = item.ProductID
	}

	products, err := s.productRepo.GetByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("order service create: %w", err)
	}

	if len(products) != len(ids) {
		return &domain.ValidationError{Field: "items.product_id", Message: "one or more products not found"}
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return fmt.Errorf("order service create: %w", err)
	}
	return nil
}

func (s *orderService) GetAll(ctx context.Context) ([]domain.Order, error) {
	orders, err := s.orderRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("order service get all: %w", err)
	}
	return orders, nil
}

func (s *orderService) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	if id == 0 {
		return nil, &domain.ValidationError{Field: "id", Message: "id is required"}
	}
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order service get by id: %w", err)
	}
	return order, nil
}

func (s *orderService) Update(ctx context.Context, order *domain.Order) error {
	if order.ID == 0 {
		return &domain.ValidationError{Field: "id", Message: "id is required"}
	}
	if order.Number == "" {
		return &domain.ValidationError{Field: "number", Message: "number is required"}
	}
	if order.DeliveryAddress == "" {
		return &domain.ValidationError{Field: "delivery_address", Message: "delivery address is required"}
	}
	if len(order.Items) == 0 {
		return &domain.ValidationError{Field: "items", Message: "order must contain at least one item"}
	}

	ids := make([]uint, len(order.Items))
	seen := make(map[uint]bool)
	for i, item := range order.Items {
		if item.Quantity <= 0 {
			return &domain.ValidationError{Field: "items.quantity", Message: "quantity must be greater than zero"}
		}
		if seen[item.ProductID] {
			return &domain.ValidationError{Field: "items.product_id", Message: fmt.Sprintf("duplicate product id %d", item.ProductID)}
		}
		seen[item.ProductID] = true
		ids[i] = item.ProductID
	}

	products, err := s.productRepo.GetByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("order service update: %w", err)
	}

	if len(products) != len(ids) {
		return &domain.ValidationError{Field: "items.product_id", Message: "one or more products not found"}
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("order service update: %w", err)
	}
	return nil
}

func (s *orderService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return &domain.ValidationError{Field: "id", Message: "id is required"}
	}

	if err := s.orderRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("order service delete: %w", err)
	}
	return nil
}