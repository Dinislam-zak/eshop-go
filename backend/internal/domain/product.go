package domain

import "time"

type Product struct {
	ID        uint
	Name      string
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}