package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Product struct {
	ID        uint
	Name      string
	Price     decimal.Decimal
	CreatedAt time.Time
	UpdatedAt time.Time
}