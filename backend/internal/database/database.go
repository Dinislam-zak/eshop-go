package database

import (
	"github.com/shop/backend/internal/config"
	"github.com/shop/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&repository.ProductModel{}, &repository.OrderModel{}, &repository.OrderItemModel{})
	if err != nil {
		return nil, err
	}

	return db, nil
}