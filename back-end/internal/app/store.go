package app

import (
	"context"
	"time"

	"github.com/srochagomes/mb-avaliacao/internal/domain"
)

type Account struct {
	ID        string
	Label     string
	CreatedAt time.Time
}

type StoredTrade struct {
	Instrument string
	CreatedAt  time.Time
	Trade      domain.Trade
}

type Tx interface {
	CreateAccount(label string, now time.Time, id string) Account
	Accounts() []Account
	Account(id string) (Account, error)

	Balance(accountID string, asset domain.Asset) *domain.Balance
	Book(instrument string) *domain.Book
	SaveOrder(o domain.Order)
	Order(id string) (domain.Order, error)
	OrdersByAccount(accountID string) []domain.Order
	AddTrade(instrument string, tr domain.Trade, at time.Time)
	Trades(instrument string) []StoredTrade
}

type Transactor interface {
	Within(ctx context.Context, fn func(Tx) error) error
}
