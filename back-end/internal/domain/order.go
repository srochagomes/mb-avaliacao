package domain

import "time"

type Side string

const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

type OrderStatus string

const (
	StatusOpen     OrderStatus = "open"
	StatusFilled   OrderStatus = "filled"
	StatusCanceled OrderStatus = "canceled"
)

type Order struct {
	ID                string
	AccountID         string
	Instrument        string // "BTC-BRL"
	Side              Side
	Price             Amount
	OriginalQuantity  Amount
	RemainingQuantity Amount
	Status            OrderStatus
	CreatedAt         time.Time
}

func (o Order) IsOpen() bool {
	return o.Status == StatusOpen && !o.RemainingQuantity.IsZero()
}

func (o Order) Clone() Order {
	o.Price = o.Price.Clone()
	o.OriginalQuantity = o.OriginalQuantity.Clone()
	o.RemainingQuantity = o.RemainingQuantity.Clone()
	return o
}
