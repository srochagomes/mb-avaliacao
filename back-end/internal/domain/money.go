package domain

import (
	"fmt"
	"math/big"
	"strings"
)

const Scale int64 = 100_000_000

type Amount struct {
	units *big.Int
}

func ParseAmount(s string) (Amount, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Amount{}, fmt.Errorf("empty amount")
	}
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return Amount{}, fmt.Errorf("invalid amount %q", s)
	}
	scale := new(big.Rat).SetInt64(Scale)
	r.Mul(r, scale)
	if !r.IsInt() {
		return Amount{}, fmt.Errorf("more than 8 decimal places: %q", s)
	}
	return Amount{units: r.Num()}, nil
}

func (a Amount) String() string {
	if a.units == nil {
		return "0"
	}
	r := new(big.Rat).SetFrac(a.units, big.NewInt(Scale))
	s := r.FloatString(8)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")

	if s == "" || s == "-" {
		return "0"
	}
	return s
}

func (a Amount) MulDiv(b Amount) Amount {
	prod := new(big.Int).Mul(a.units, b.units)
	prod.Div(prod, big.NewInt(Scale))
	return Amount{units: prod}
}

func unitsOrZero(a Amount) *big.Int {
	if a.units == nil {
		return new(big.Int)
	}
	return a.units
}

func (a Amount) Add(b Amount) Amount {
	return Amount{units: new(big.Int).Add(unitsOrZero(a), unitsOrZero(b))}
}

func (a Amount) Sub(b Amount) (Amount, error) {
	diff := new(big.Int).Sub(unitsOrZero(a), unitsOrZero(b))
	if diff.Sign() < 0 {
		return Amount{}, fmt.Errorf("insufficient amount")
	}
	return Amount{units: diff}, nil
}

func (a Amount) Cmp(b Amount) int {
	au, bu := a.units, b.units
	if au == nil {
		au = new(big.Int)
	}
	if bu == nil {
		bu = new(big.Int)
	}
	return au.Cmp(bu)
}

func (a Amount) IsZero() bool {
	return a.units == nil || a.units.Sign() == 0
}

func (a Amount) Clone() Amount {
	if a.units == nil || a.units.Sign() == 0 {
		return Amount{}
	}
	return Amount{units: new(big.Int).Set(a.units)}
}
