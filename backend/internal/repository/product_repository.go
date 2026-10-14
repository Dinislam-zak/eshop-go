package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shop/backend/internal/domain"
	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := ProductModel{
			Name:  product.Name,
			Price: product.Price,
		}
		if err := tx.Create(&model).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrAlreadyExists
			}
			return err
		}
		product.ID = model.ID
		product.CreatedAt = model.CreatedAt
		product.UpdatedAt = model.UpdatedAt
		return nil
	})
}

func (r *productRepository) GetAll(ctx context.Context) ([]domain.Product, error) {
	var models []ProductModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	products := make([]domain.Product, len(models))
	for i, m := range models {
		products[i] = toProductDomain(&m)
	}
	return products, nil
}

func (r *productRepository) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	var model ProductModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	product := toProductDomain(&model)
	return &product, nil
}

func (r *productRepository) GetByName(ctx context.Context, name string) (*domain.Product, error) {
	var model ProductModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	product := toProductDomain(&model)
	return &product, nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing ProductModel
		if err := tx.First(&existing, product.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		existing.Name = product.Name
		existing.Price = product.Price

		if err := tx.Save(&existing).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrAlreadyExists
			}
			return err
		}
		product.UpdatedAt = existing.UpdatedAt
		return nil
	})
}

func (r *productRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&ProductModel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *productRepository) ExistsInOrders(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&OrderItemModel{}).Where("product_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func toProductDomain(m *ProductModel) domain.Product {
	return domain.Product{
		ID:        m.ID,
		Name:      m.Name,
		Price:     m.Price,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}