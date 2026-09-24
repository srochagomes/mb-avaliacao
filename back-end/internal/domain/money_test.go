package domain

import "testing"

func TestParseAmount(t *testing.T) {
	tests := []struct {
		in   string
		want string // String() esperado
	}{
		{"1", "1"},
		{"500000", "500000"},
		{"0.00000001", "0.00000001"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) { 
			/* Parse + String */ 
			got, err := ParseAmount(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != tt.want {
				t.Fatalf("got %q want %q", got.String(), tt.want)
			}})
	}
}

func TestAmountAddSub(t *testing.T) {
	a, _ := ParseAmount("10")
	b, _ := ParseAmount("3")
	if a.Add(b).String() != "13" {
		t.Fatalf("Add: got %s", a.Add(b).String())
	}
	got, err := a.Sub(b)
	if err != nil || got.String() != "7" {
		t.Fatalf("Sub: got %s err=%v", got.String(), err)
	}
	_, err = b.Sub(a) // 3 - 10
	if err == nil {
		t.Fatal("Sub underflow must error")
	}
}

func TestAmountCmpIsZero(t *testing.T) {
	a, _ := ParseAmount("1")
	b, _ := ParseAmount("2")
	z, _ := ParseAmount("0")
	if a.Cmp(b) >= 0 {
		t.Fatal("1 < 2")
	}
	if !z.IsZero() {
		t.Fatal("0 must be zero")
	}
}
func TestMulDivNotional(t *testing.T) {
	price, _ := ParseAmount("500000")
	qty, _ := ParseAmount("1")
	if price.MulDiv(qty).String() != "500000" {
		t.Fatal(price.MulDiv(qty).String())
	}
	// bônus: 510000 * 1 = 510000 (reserva cheia antes do match)
}

func TestNotionalExample(t *testing.T) {
	// preço 500000 * qty 1 / scale = 500000
}