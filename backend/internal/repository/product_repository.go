package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shop/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *productRepository {
	return &productRepository{db: db}
}

func (r *productRepository) getDB(ctx context.Context) *gorm.DB {
	if tx := getTx(ctx); tx != nil {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	db := r.getDB(ctx)
	model := ProductModel{
		Name:  product.Name,
		Price: product.Price,
	}
	if err := db.Create(&model).Error; err != nil {
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
}

func (r *productRepository) GetAll(ctx context.Context, page, pageSize int) ([]domain.Product, error) {
	var models []ProductModel
	offset := (page - 1) * pageSize
	if err := r.getDB(ctx).Offset(offset).Limit(pageSize).Find(&models).Error; err != nil {
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
	if err := r.getDB(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	product := toProductDomain(&model)
	return &product, nil
}

func (r *productRepository) GetByIDs(ctx context.Context, ids []uint) ([]domain.Product, error) {
	var models []ProductModel
	if err := r.getDB(ctx).Where("id IN ?", ids).Find(&models).Error; err != nil {
		return nil, err
	}

	products := make([]domain.Product, len(models))
	for i, m := range models {
		products[i] = toProductDomain(&m)
	}
	return products, nil
}

func (r *productRepository) GetByName(ctx context.Context, name string) (*domain.Product, error) {
	var model ProductModel
	if err := r.getDB(ctx).Where("name = ?", name).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	product := toProductDomain(&model)
	return &product, nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	db := r.getDB(ctx)
	var existing ProductModel
	if err := db.First(&existing, product.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	existing.Name = product.Name
	existing.Price = product.Price

	if err := db.Save(&existing).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyExists
		}
		return err
	}
	product.UpdatedAt = existing.UpdatedAt
	return nil
}

func (r *productRepository) Delete(ctx context.Context, id uint) error {
	db := r.getDB(ctx)
	var existing ProductModel
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	var count int64
	if err := db.Model(&OrderItemModel{}).Where("product_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrInUse
	}

	res := db.Delete(&existing)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *productRepository) ExistsInOrders(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.getDB(ctx).Model(&OrderItemModel{}).Where("product_id = ?", id).Count(&count).Error
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