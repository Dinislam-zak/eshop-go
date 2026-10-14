package service

import (
	"context"
	"fmt"

	"github.com/shop/backend/internal/domain"
	"github.com/shop/backend/internal/repository"
)

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) Create(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	if product.Name == "" {
		return nil, &domain.ValidationError{Field: "name", Message: "name is required"}
	}
	if product.Price <= 0 {
		return nil, &domain.ValidationError{Field: "price", Message: "price must be greater than zero"}
	}
	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("product service create: %w", err)
	}
	created, err := s.repo.GetByID(ctx, product.ID)
	if err != nil {
		return nil, fmt.Errorf("product service create: get created: %w", err)
	}
	return created, nil
}

func (s *productService) GetAll(ctx context.Context) ([]domain.Product, error) {
	products, err := s.repo.GetAll(ctx)
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
	if product.Price <= 0 {
		return nil, &domain.ValidationError{Field: "price", Message: "price must be greater than zero"}
	}
	if err := s.repo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("product service update: %w", err)
	}
	updated, err := s.repo.GetByID(ctx, product.ID)
	if err != nil {
		return nil, fmt.Errorf("product service update: get updated: %w", err)
	}
	return updated, nil
}

func (s *productService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return &domain.ValidationError{Field: "id", Message: "id is required"}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("product service delete: %w", err)
	}
	return nil
}