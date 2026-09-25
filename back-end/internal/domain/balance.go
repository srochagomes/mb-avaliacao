package domain

import "errors"

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
)

type Asset string
const (
	AssetBTC Asset = "BTC"
	AssetBRL Asset = "BRL"
)

type Balance struct {
	AccountID string
	Asset     Asset
	available Amount
	reserved  Amount
}
func NewBalance(accountID string, asset Asset) Balance {
	return Balance{AccountID: accountID, Asset: asset}
}
func (b Balance) Available() Amount { return b.available }
func (b Balance) Reserved() Amount  { return b.reserved }




func (b *Balance) Credit(amount Amount) error {
	if amount.Cmp(Amount{}) <= 0 {
		return ErrInvalidAmount
	}
	b.available = b.available.Add(amount)
	return nil
}

func (b *Balance) Debit(amount Amount) error {
	if amount.Cmp(Amount{}) <= 0 {
		return ErrInvalidAmount
	}
	next, err := b.available.Sub(amount)
	if err != nil {
		return ErrInsufficientBalance
	}
	b.available = next
	return nil
}

func (b *Balance) Reserve(amount Amount) error {
	if amount.Cmp(Amount{}) <= 0 {
		return ErrInvalidAmount
	}
	next, err := b.available.Sub(amount)
	if err != nil {
		return ErrInsufficientBalance
	}
	b.available = next
	b.reserved = b.reserved.Add(amount)
	return nil
}

func (b *Balance) Release(amount Amount) error {
	if amount.Cmp(Amount{}) <= 0 {
		return ErrInvalidAmount
	}
	next, err := b.reserved.Sub(amount)
	if err != nil {
		return ErrInsufficientBalance
	}
	b.reserved = next
	b.available = b.available.Add(amount)
	return nil
}

func (b *Balance) ConsumeReserved(amount Amount) error {
	if amount.Cmp(Amount{}) <= 0 {
		return ErrInvalidAmount
	}
	next, err := b.reserved.Sub(amount)
	if err != nil {
		return ErrInsufficientBalance
	}
	b.reserved = next
	return nil
}