package repository

import (
	"time"

	"gorm.io/gorm"
)

type ProductModel struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string         `gorm:"uniqueIndex:idx_product_name_active,where:deleted_at IS NULL;not null"`
	Price     float64        `gorm:"not null"`
}

type OrderModel struct {
	ID              uint             `gorm:"primarykey"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt   `gorm:"index"`
	Number          string           `gorm:"uniqueIndex:idx_order_number_active,where:deleted_at IS NULL;not null"`
	DeliveryAddress string           `gorm:"not null"`
	Items           []OrderItemModel `gorm:"foreignKey:OrderID"`
}

type OrderItemModel struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	OrderID   uint           `gorm:"not null;index"`
	ProductID uint           `gorm:"not null;index"`
	Quantity  int            `gorm:"not null"`
	Product   ProductModel   `gorm:"foreignKey:ProductID"`
}