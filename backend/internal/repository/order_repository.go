package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shop/backend/internal/domain"
	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *orderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var items []OrderItemModel
		for _, item := range order.Items {
			items = append(items, OrderItemModel{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}

		model := OrderModel{
			Number:          order.Number,
			DeliveryAddress: order.DeliveryAddress,
			Items:           items,
		}

		if err := tx.Create(&model).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrAlreadyExists
			}
			return err
		}
		order.ID = model.ID
		order.CreatedAt = model.CreatedAt
		order.UpdatedAt = model.UpdatedAt
		return nil
	})
}

func (r *orderRepository) GetAll(ctx context.Context) ([]domain.Order, error) {
	var models []OrderModel
	if err := r.db.WithContext(ctx).Preload("Items.Product").Find(&models).Error; err != nil {
		return nil, err
	}

	orders := make([]domain.Order, len(models))
	for i, m := range models {
		orders[i] = toOrderDomain(&m)
	}
	return orders, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	var model OrderModel
	if err := r.db.WithContext(ctx).Preload("Items.Product").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	order := toOrderDomain(&model)
	return &order, nil
}

func (r *orderRepository) GetByNumber(ctx context.Context, number string) (*domain.Order, error) {
	var model OrderModel
	if err := r.db.WithContext(ctx).Where("number = ?", number).Preload("Items.Product").First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	order := toOrderDomain(&model)
	return &order, nil
}

func (r *orderRepository) Update(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing OrderModel
		if err := tx.Preload("Items").First(&existing, order.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		existing.Number = order.Number
		existing.DeliveryAddress = order.DeliveryAddress
		if err := tx.Save(&existing).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrAlreadyExists
			}
			return err
		}

		if err := tx.Where("order_id = ?", existing.ID).Delete(&OrderItemModel{}).Error; err != nil {
			return err
		}

		var items []OrderItemModel
		for _, item := range order.Items {
			items = append(items, OrderItemModel{
				OrderID:   existing.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}

		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		order.UpdatedAt = existing.UpdatedAt
		return nil
	})
}

func (r *orderRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", id).Delete(&OrderItemModel{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&OrderModel{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
}

func toOrderDomain(m *OrderModel) domain.Order {
	items := make([]domain.OrderItem, len(m.Items))
	for i, item := range m.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Product:   toProductDomain(&item.Product),
			Quantity:  item.Quantity,
		}
	}
	return domain.Order{
		ID:              m.ID,
		Number:          m.Number,
		DeliveryAddress: m.DeliveryAddress,
		Items:           items,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}