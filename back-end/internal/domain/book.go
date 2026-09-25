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

func (b Book) Bids() []Order { return cloneOrders(b.bids) }
func (b Book) Asks() []Order { return cloneOrders(b.asks) }

func (b Book) Clone() Book {
	return Book{
		Instrument: b.Instrument,
		bids:       cloneOrders(b.bids),
		asks:       cloneOrders(b.asks),
	}
}

func cloneOrders(in []Order) []Order {
	if len(in) == 0 {
		return nil
	}
	out := make([]Order, len(in))
	for i := range in {
		out[i] = in[i].Clone()
	}
	return out
}

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

func (b *Book) Remove(id string) (Order, error) {
	if o, rest, ok := take(b.bids, id); ok {
		b.bids = rest
		return o, nil
	}
	if o, rest, ok := take(b.asks, id); ok {
		b.asks = rest
		return o, nil
	}
	return Order{}, ErrOrderNotOpen
}

func take(orders []Order, id string) (Order, []Order, bool) {
	for i, o := range orders {
		if o.ID == id {
			return o, slices.Delete(orders, i, i+1), true
		}
	}
	return Order{}, orders, false
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
