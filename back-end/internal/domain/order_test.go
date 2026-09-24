package domain

import "testing"

func TestOrderIsOpen(t *testing.T) {
	qty, _ := ParseAmount("1")
	o := Order{Status: StatusOpen, RemainingQuantity: qty}
	if !o.IsOpen() {
		t.Fatal("expected open")
	}
	o.Status = StatusCanceled
	if o.IsOpen() {
		t.Fatal("canceled must not be open")
	}
}