package domain

import (
	"testing"
	"time"
)

func TestMatch(t *testing.T) {
	at := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)

	t.Run("full", func(t *testing.T) {
		book := NewBook("BTC-BRL")
		maker := openOrder(t, "maker", SideSell, "500000", at)
		maker.AccountID = "B"
		if err := book.Add(maker); err != nil {
			t.Fatal(err)
		}
		taker := openOrder(t, "taker", SideBuy, "500000", at.Add(time.Second))
		taker.AccountID = "A"

		trades, got, err := book.Match(taker)
		if err != nil {
			t.Fatal(err)
		}
		if len(trades) != 1 || trades[0].Price.String() != "500000" || trades[0].MakerOrderID != "maker" {
			t.Fatalf("trades %+v", trades)
		}
		if got.Status != StatusFilled || len(book.Bids()) != 0 || len(book.Asks()) != 0 {
			t.Fatalf("status %s bids %d asks %d", got.Status, len(book.Bids()), len(book.Asks()))
		}
	})

	t.Run("partial", func(t *testing.T) {
		book := NewBook("BTC-BRL")
		maker := openOrder(t, "maker", SideSell, "500000", at)
		maker.AccountID = "B"
		if err := book.Add(maker); err != nil {
			t.Fatal(err)
		}
		taker := openOrder(t, "taker", SideBuy, "500000", at.Add(time.Second))
		taker.AccountID = "A"
		taker.OriginalQuantity = mustAmount(t, "2")
		taker.RemainingQuantity = mustAmount(t, "2")

		trades, got, err := book.Match(taker)
		if err != nil {
			t.Fatal(err)
		}
		if len(trades) != 1 || trades[0].Quantity.String() != "1" {
			t.Fatalf("trades %+v", trades)
		}
		if got.Status != StatusOpen || got.RemainingQuantity.String() != "1" || len(book.Bids()) != 1 || len(book.Asks()) != 0 {
			t.Fatalf("status %s rem %s bids %d asks %d", got.Status, got.RemainingQuantity, len(book.Bids()), len(book.Asks()))
		}
	})

	t.Run("maker price", func(t *testing.T) {
		book := NewBook("BTC-BRL")
		maker := openOrder(t, "maker", SideSell, "500000", at)
		maker.AccountID = "B"
		if err := book.Add(maker); err != nil {
			t.Fatal(err)
		}
		taker := openOrder(t, "taker", SideBuy, "510000", at.Add(time.Second))
		taker.AccountID = "A"

		trades, _, err := book.Match(taker)
		if err != nil {
			t.Fatal(err)
		}
		if len(trades) != 1 || trades[0].Price.String() != "500000" {
			t.Fatalf("trades %+v", trades)
		}
	})

	t.Run("no cross", func(t *testing.T) {
		book := NewBook("BTC-BRL")
		maker := openOrder(t, "maker", SideSell, "101", at)
		maker.AccountID = "B"
		if err := book.Add(maker); err != nil {
			t.Fatal(err)
		}
		taker := openOrder(t, "taker", SideBuy, "100", at.Add(time.Second))
		taker.AccountID = "A"

		trades, got, err := book.Match(taker)
		if err != nil {
			t.Fatal(err)
		}
		if len(trades) != 0 || got.Status != StatusOpen || len(book.Bids()) != 1 || book.Asks()[0].ID != "maker" {
			t.Fatalf("trades %d status %s bids %d", len(trades), got.Status, len(book.Bids()))
		}
	})

	t.Run("self trade", func(t *testing.T) {
		book := NewBook("BTC-BRL")
		maker := openOrder(t, "maker", SideSell, "100", at)
		maker.AccountID = "A"
		if err := book.Add(maker); err != nil {
			t.Fatal(err)
		}
		taker := openOrder(t, "taker", SideBuy, "100", at.Add(time.Second))
		taker.AccountID = "A"

		trades, _, err := book.Match(taker)
		if err != nil {
			t.Fatal(err)
		}
		if len(trades) != 0 || len(book.Bids()) != 1 || len(book.Asks()) != 1 {
			t.Fatalf("trades %d bids %d asks %d", len(trades), len(book.Bids()), len(book.Asks()))
		}
	})

	t.Run("fifo", func(t *testing.T) {
		book := NewBook("BTC-BRL")
		late := openOrder(t, "late", SideSell, "100", at.Add(time.Second))
		late.AccountID = "B"
		early := openOrder(t, "early", SideSell, "100", at)
		early.AccountID = "C"
		if err := book.Add(late); err != nil {
			t.Fatal(err)
		}
		if err := book.Add(early); err != nil {
			t.Fatal(err)
		}
		taker := openOrder(t, "taker", SideBuy, "100", at.Add(2*time.Second))
		taker.AccountID = "A"

		trades, _, err := book.Match(taker)
		if err != nil {
			t.Fatal(err)
		}
		if len(trades) != 1 || trades[0].MakerOrderID != "early" || len(book.Asks()) != 1 || book.Asks()[0].ID != "late" {
			t.Fatalf("trades %+v asks %v", trades, book.Asks())
		}
	})
}
