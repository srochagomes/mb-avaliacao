package domain

import (
	"errors"
	"testing"
	"time"
)

func TestSettle(t *testing.T) {
	aBRL := NewBalance("A", AssetBRL)
	aBTC := NewBalance("A", AssetBTC)
	bBRL := NewBalance("B", AssetBRL)
	bBTC := NewBalance("B", AssetBTC)
	if err := aBRL.Credit(mustAmount(t, "510000")); err != nil {
		t.Fatal(err)
	}
	if err := bBTC.Credit(mustAmount(t, "1")); err != nil {
		t.Fatal(err)
	}

	at := mustTime(t)
	book := NewBook("BTC-BRL")
	sell := openOrder(t, "sell", SideSell, "500000", at)
	sell.AccountID = "B"
	if err := ReserveForOrder(&bBTC, sell); err != nil {
		t.Fatal(err)
	}
	if err := book.Add(sell); err != nil {
		t.Fatal(err)
	}

	buy := openOrder(t, "buy", SideBuy, "510000", at)
	buy.AccountID = "A"
	if err := ReserveForOrder(&aBRL, buy); err != nil {
		t.Fatal(err)
	}
	trades, _, err := book.Match(buy)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 {
		t.Fatalf("trades %d", len(trades))
	}
	if err := Settle(&aBRL, &aBTC, &bBRL, &bBTC, trades[0], buy.Price); err != nil {
		t.Fatal(err)
	}

	if aBTC.Available().String() != "1" || aBRL.Available().String() != "10000" || aBRL.Reserved().String() != "0" {
		t.Fatalf("A btc %s brl %s reserved %s", aBTC.Available(), aBRL.Available(), aBRL.Reserved())
	}
	if bBTC.Available().String() != "0" || bBTC.Reserved().String() != "0" || bBRL.Available().String() != "500000" {
		t.Fatalf("B btc %s reserved %s brl %s", bBTC.Available(), bBTC.Reserved(), bBRL.Available())
	}
}

func mustTime(t *testing.T) time.Time {
	t.Helper()
	return time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
}

func TestSettleSamePrice(t *testing.T) {
	aBRL := NewBalance("A", AssetBRL)
	aBTC := NewBalance("A", AssetBTC)
	bBRL := NewBalance("B", AssetBRL)
	bBTC := NewBalance("B", AssetBTC)
	if err := aBRL.Credit(mustAmount(t, "500000")); err != nil {
		t.Fatal(err)
	}
	if err := bBTC.Credit(mustAmount(t, "1")); err != nil {
		t.Fatal(err)
	}

	at := mustTime(t)
	book := NewBook("BTC-BRL")
	buy := openOrder(t, "buy", SideBuy, "500000", at)
	buy.AccountID = "A"
	if err := ReserveForOrder(&aBRL, buy); err != nil {
		t.Fatal(err)
	}
	if err := book.Add(buy); err != nil {
		t.Fatal(err)
	}

	sell := openOrder(t, "sell", SideSell, "500000", at)
	sell.AccountID = "B"
	if err := ReserveForOrder(&bBTC, sell); err != nil {
		t.Fatal(err)
	}
	trades, _, err := book.Match(sell)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 {
		t.Fatalf("trades %d", len(trades))
	}
	if err := Settle(&aBRL, &aBTC, &bBRL, &bBTC, trades[0], buy.Price); err != nil {
		t.Fatal(err)
	}

	if aBTC.Available().String() != "1" || aBRL.Available().String() != "0" || aBRL.Reserved().String() != "0" {
		t.Fatalf("A btc %s brl %s reserved %s", aBTC.Available(), aBRL.Available(), aBRL.Reserved())
	}
	if bBTC.Available().String() != "0" || bBTC.Reserved().String() != "0" || bBRL.Available().String() != "500000" {
		t.Fatalf("B btc %s reserved %s brl %s", bBTC.Available(), bBTC.Reserved(), bBRL.Available())
	}
}

func TestReleaseForOrder(t *testing.T) {
	bal := NewBalance("A", AssetBRL)
	if err := bal.Credit(mustAmount(t, "100")); err != nil {
		t.Fatal(err)
	}
	o := openOrder(t, "o", SideBuy, "40", mustTime(t))
	o.AccountID = "A"
	if err := ReserveForOrder(&bal, o); err != nil {
		t.Fatal(err)
	}
	if err := ReleaseForOrder(&bal, o); err != nil {
		t.Fatal(err)
	}
	if bal.Available().String() != "100" || bal.Reserved().String() != "0" {
		t.Fatalf("available %s reserved %s", bal.Available(), bal.Reserved())
	}

	o.Status = StatusFilled
	if err := ReleaseForOrder(&bal, o); !errors.Is(err, ErrOrderNotOpen) {
		t.Fatalf("filled: %v", err)
	}
}
