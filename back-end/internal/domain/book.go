package domain

import (
	"errors"
	"slices"
)

var (
	ErrOrderNotOpen       = errors.New("order not open")
	ErrInstrumentMismatch = errors.New("instrument mismatch")
)

type Book struct {
	Instrument string
	bids       []Order
	asks       []Order
}

func NewBook(instrument string) Book {
	return Book{Instrument: instrument}
}

func (b Book) Bids() []Order { return slices.Clone(b.bids) }
func (b Book) Asks() []Order { return slices.Clone(b.asks) }

func (b *Book) Add(o Order) error {
	if o.Instrument != b.Instrument {
		return ErrInstrumentMismatch
	}
	if !o.IsOpen() {
		return ErrOrderNotOpen
	}
	if o.Side == SideBuy {
		b.bids = insert(b.bids, o, bidBefore)
		return nil
	}
	if o.Side == SideSell {
		b.asks = insert(b.asks, o, askBefore)
		return nil
	}
	return ErrInvalidAmount
}

func bidBefore(a, b Order) bool {
	cmp := a.Price.Cmp(b.Price)
	if cmp != 0 {
		return cmp > 0
	}
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.Before(b.CreatedAt)
	}
	return a.ID < b.ID
}

func askBefore(a, b Order) bool {
	cmp := a.Price.Cmp(b.Price)
	if cmp != 0 {
		return cmp < 0
	}
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.Before(b.CreatedAt)
	}
	return a.ID < b.ID
}
func insert(orders []Order, o Order, before func(a, b Order) bool) []Order {
	i := 0
	for i < len(orders) && !before(o, orders[i]) {
		i++
	}
	orders = append(orders, Order{})
	copy(orders[i+1:], orders[i:])
	orders[i] = o
	return orders
}
