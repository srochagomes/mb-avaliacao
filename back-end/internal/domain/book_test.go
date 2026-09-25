package domain

import (
	"errors"
	"testing"
	"time"
)

func openOrder(t *testing.T, id string, side Side, price string, at time.Time) Order {
	t.Helper()
	qty := mustAmount(t, "1")
	return Order{
		ID: id, AccountID: "acc", Instrument: "BTC-BRL", Side: side,
		Price: mustAmount(t, price), OriginalQuantity: qty, RemainingQuantity: qty,
		Status: StatusOpen, CreatedAt: at,
	}
}

func TestBook(t *testing.T) {
	at := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	book := NewBook("BTC-BRL")

	if err := book.Add(openOrder(t, "low", SideBuy, "100", at)); err != nil {
		t.Fatal(err)
	}
	if err := book.Add(openOrder(t, "high", SideBuy, "101", at)); err != nil {
		t.Fatal(err)
	}
	if book.Bids()[0].ID != "high" || len(book.Asks()) != 0 {
		t.Fatalf("bids %s asks %d", book.Bids()[0].ID, len(book.Asks()))
	}

	fifo := NewBook("BTC-BRL")
	fifo.Add(openOrder(t, "late", SideBuy, "100", at.Add(time.Second)))
	fifo.Add(openOrder(t, "early", SideBuy, "100", at))
	if fifo.Bids()[0].ID != "early" {
		t.Fatalf("time: %s", fifo.Bids()[0].ID)
	}

	sameTime := NewBook("BTC-BRL")
	sameTime.Add(openOrder(t, "b", SideBuy, "100", at))
	sameTime.Add(openOrder(t, "a", SideBuy, "100", at))
	if sameTime.Bids()[0].ID != "a" {
		t.Fatalf("id: %s", sameTime.Bids()[0].ID)
	}

	asks := NewBook("BTC-BRL")
	asks.Add(openOrder(t, "s100", SideSell, "100", at))
	asks.Add(openOrder(t, "s99", SideSell, "99", at))
	if asks.Asks()[0].ID != "s99" {
		t.Fatalf("asks: %s", asks.Asks()[0].ID)
	}

	canceled := openOrder(t, "c", SideBuy, "100", at)
	canceled.Status = StatusCanceled
	if err := book.Add(canceled); !errors.Is(err, ErrOrderNotOpen) {
		t.Fatalf("canceled: %v", err)
	}

	other := openOrder(t, "e", SideBuy, "100", at)
	other.Instrument = "ETH-BRL"
	if err := book.Add(other); !errors.Is(err, ErrInstrumentMismatch) {
		t.Fatalf("instrument: %v", err)
	}
}
