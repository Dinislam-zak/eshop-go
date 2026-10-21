package service

import (
	"context"
	"fmt"

	"github.com/shop/backend/internal/domain"
)

type orderService struct {
	orderRepo   OrderRepository
	productRepo ProductRepository
	tx          TransactionManager
}

func NewOrderService(orderRepo OrderRepository, productRepo ProductRepository, tx TransactionManager) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		tx:          tx,
	}
}

func (s *orderService) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	if order.Number == "" {
		return nil, &domain.ValidationError{Field: "number", Message: "number is required"}
	}
	if order.DeliveryAddress == "" {
		return nil, &domain.ValidationError{Field: "delivery_address", Message: "delivery address is required"}
	}
	if len(order.Items) == 0 {
		return nil, &domain.ValidationError{Field: "items", Message: "order must contain at least one item"}
	}

	ids := make([]uint, len(order.Items))
	seen := make(map[uint]bool)
	for i, item := range order.Items {
		if item.Quantity <= 0 {
			return nil, &domain.ValidationError{Field: "items.quantity", Message: "quantity must be greater than zero"}
		}
		if seen[item.ProductID] {
			return nil, &domain.ValidationError{Field: "items.product_id", Message: fmt.Sprintf("duplicate product id %d", item.ProductID)}
		}
		seen[item.ProductID] = true
		ids[i] = item.ProductID
	}

	var created *domain.Order
	err := s.tx.Execute(ctx, func(txCtx context.Context) error {
		products, err := s.productRepo.GetByIDs(txCtx, ids)
		if err != nil {
			return fmt.Errorf("order service create: %w", err)
		}
		if len(products) != len(ids) {
			return &domain.ValidationError{Field: "items.product_id", Message: "one or more products not found"}
		}

		if err := s.orderRepo.Create(txCtx, order); err != nil {
			return fmt.Errorf("order service create: %w", err)
		}
		
		createdOrder, err := s.orderRepo.GetByID(txCtx, order.ID)
		if err != nil {
			return fmt.Errorf("order service create: %w", err)
		}
		created = createdOrder
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *orderService) GetAll(ctx context.Context, page, pageSize int) ([]domain.Order, error) {
	orders, err := s.orderRepo.GetAll(ctx, page, pageSize)
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

func (s *orderService) Update(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	if order.ID == 0 {
		return nil, &domain.ValidationError{Field: "id", Message: "id is required"}
	}
	if order.Number == "" {
		return nil, &domain.ValidationError{Field: "number", Message: "number is required"}
	}
	if order.DeliveryAddress == "" {
		return nil, &domain.ValidationError{Field: "delivery_address", Message: "delivery address is required"}
	}
	if len(order.Items) == 0 {
		return nil, &domain.ValidationError{Field: "items", Message: "order must contain at least one item"}
	}

	ids := make([]uint, len(order.Items))
	seen := make(map[uint]bool)
	for i, item := range order.Items {
		if item.Quantity <= 0 {
			return nil, &domain.ValidationError{Field: "items.quantity", Message: "quantity must be greater than zero"}
		}
		if seen[item.ProductID] {
			return nil, &domain.ValidationError{Field: "items.product_id", Message: fmt.Sprintf("duplicate product id %d", item.ProductID)}
		}
		seen[item.ProductID] = true
		ids[i] = item.ProductID
	}

	var updated *domain.Order
	err := s.tx.Execute(ctx, func(txCtx context.Context) error {
		products, err := s.productRepo.GetByIDs(txCtx, ids)
		if err != nil {
			return fmt.Errorf("order service update: %w", err)
		}
		if len(products) != len(ids) {
			return &domain.ValidationError{Field: "items.product_id", Message: "one or more products not found"}
		}

		if err := s.orderRepo.Update(txCtx, order); err != nil {
			return fmt.Errorf("order service update: %w", err)
		}
		
		updatedOrder, err := s.orderRepo.GetByID(txCtx, order.ID)
		if err != nil {
			return fmt.Errorf("order service update: %w", err)
		}
		updated = updatedOrder
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *orderService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return &domain.ValidationError{Field: "id", Message: "id is required"}
	}
	return s.tx.Execute(ctx, func(txCtx context.Context) error {
		if err := s.orderRepo.Delete(txCtx, id); err != nil {
			return fmt.Errorf("order service delete: %w", err)
		}
		return nil
	})
}