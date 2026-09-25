package domain

import (
	"errors"
	"testing"
)

func mustAmount(t *testing.T, s string) Amount {
	t.Helper()
	a, err := ParseAmount(s)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestBalance(t *testing.T) {
	b := NewBalance("acc-1", AssetBRL)
	if err := b.Credit(mustAmount(t, "100")); err != nil {
		t.Fatal(err)
	}
	if err := b.Reserve(mustAmount(t, "40")); err != nil {
		t.Fatal(err)
	}
	if b.Available().String() != "60" || b.Reserved().String() != "40" {
		t.Fatalf("available %s reserved %s", b.Available(), b.Reserved())
	}

	if err := b.Reserve(mustAmount(t, "70")); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("reserve: %v", err)
	}
	if b.Available().String() != "60" || b.Reserved().String() != "40" {
		t.Fatalf("saldo mudou: available %s reserved %s", b.Available(), b.Reserved())
	}
	if err := b.Debit(mustAmount(t, "60")); err != nil {
		t.Fatal(err)
	}
	if err := b.Debit(mustAmount(t, "1")); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("debit: %v", err)
	}
	if b.Reserved().String() != "40" {
		t.Fatalf("reserved %s", b.Reserved())
	}

	if err := b.Release(mustAmount(t, "40")); err != nil {
		t.Fatal(err)
	}
	if b.Available().String() != "40" || b.Reserved().String() != "0" {
		t.Fatalf("available %s reserved %s", b.Available(), b.Reserved())
	}

	other := NewBalance("acc-2", AssetBTC)
	if err := other.Credit(mustAmount(t, "50")); err != nil {
		t.Fatal(err)
	}
	if err := other.Reserve(mustAmount(t, "50")); err != nil {
		t.Fatal(err)
	}
	if err := other.ConsumeReserved(mustAmount(t, "20")); err != nil {
		t.Fatal(err)
	}
	if other.Available().String() != "0" || other.Reserved().String() != "30" {
		t.Fatalf("available %s reserved %s", other.Available(), other.Reserved())
	}

	if err := b.Credit(mustAmount(t, "0")); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("credit zero: %v", err)
	}
	if err := b.Credit(mustAmount(t, "-1")); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("credit negative: %v", err)
	}
}
