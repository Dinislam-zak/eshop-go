package domain

import "time"

type OrderItem struct {
	ProductID uint
	Product   Product
	Quantity  int
}

type Order struct {
	ID              uint
	Number          string
	DeliveryAddress string
	Items           []OrderItem
	CreatedAt       time.Time
	UpdatedAt       time.Time
}