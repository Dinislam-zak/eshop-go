package service

import (
	"context"
	"fmt"

	"github.com/shop/backend/internal/domain"
)

type productService struct {
	repo ProductRepository
	tx   TransactionManager
}

func NewProductService(repo ProductRepository, tx TransactionManager) ProductService {
	return &productService{repo: repo, tx: tx}
}

func (s *productService) Create(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	if product.Name == "" {
		return nil, &domain.ValidationError{Field: "name", Message: "name is required"}
	}
	if product.Price.IsNegative() || product.Price.IsZero() {
		return nil, &domain.ValidationError{Field: "price", Message: "price must be greater than zero"}
	}
	var created *domain.Product
	err := s.tx.Execute(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, product); err != nil {
			return fmt.Errorf("product service create: %w", err)
		}
		p, err := s.repo.GetByID(txCtx, product.ID)
		if err != nil {
			return fmt.Errorf("product service create: %w", err)
		}
		created = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *productService) GetAll(ctx context.Context, page, pageSize int) ([]domain.Product, error) {
	products, err := s.repo.GetAll(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("product service get all: %w", err)
	}
	return products, nil
}

func (s *productService) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	if id == 0 {
		return nil, &domain.ValidationError{Field: "id", Message: "id is required"}
	}
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("product service get by id: %w", err)
	}
	return product, nil
}

func (s *productService) Update(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	if product.ID == 0 {
		return nil, &domain.ValidationError{Field: "id", Message: "id is required"}
	}
	if product.Name == "" {
		return nil, &domain.ValidationError{Field: "name", Message: "name is required"}
	}
	if product.Price.IsNegative() || product.Price.IsZero() {
		return nil, &domain.ValidationError{Field: "price", Message: "price must be greater than zero"}
	}
	var updated *domain.Product
	err := s.tx.Execute(ctx, func(txCtx context.Context) error {
		if err := s.repo.Update(txCtx, product); err != nil {
			return fmt.Errorf("product service update: %w", err)
		}
		p, err := s.repo.GetByID(txCtx, product.ID)
		if err != nil {
			return fmt.Errorf("product service update: %w", err)
		}
		updated = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *productService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return &domain.ValidationError{Field: "id", Message: "id is required"}
	}
	return s.tx.Execute(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, id); err != nil {
			return fmt.Errorf("product service delete: %w", err)
		}
		return nil
	})
}