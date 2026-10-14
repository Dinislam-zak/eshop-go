package repository

import (
	"context"
	"errors"

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
	var existing ProductModel
	if err := r.db.WithContext(ctx).Where("name = ?", product.Name).First(&existing).Error; err == nil {
		return domain.ErrAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	model := ProductModel{
		Name:  product.Name,
		Price: product.Price,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	product.ID = model.ID
	return nil
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
	var existing ProductModel
	if err := r.db.WithContext(ctx).First(&existing, product.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&ProductModel{}).Where("name = ? AND id != ?", product.Name, product.ID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrAlreadyExists
	}

	existing.Name = product.Name
	existing.Price = product.Price

	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return err
	}
	return nil
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
		ID:    m.ID,
		Name:  m.Name,
		Price: m.Price,
	}
}