package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/srochagomes/mb-avaliacao/internal/app"
	"github.com/srochagomes/mb-avaliacao/internal/domain"
)

type Store struct {
	mu       sync.Mutex
	accounts map[string]app.Account
	balances map[string]*domain.Balance
	books    map[string]*domain.Book
	orders   map[string]domain.Order
	trades   []app.StoredTrade
}

func New() *Store {
	return &Store{
		accounts: map[string]app.Account{},
		balances: map[string]*domain.Balance{},
		books:    map[string]*domain.Book{},
		orders:   map[string]domain.Order{},
	}
}

func (s *Store) Within(ctx context.Context, fn func(app.Tx) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.snapshot()
	if err := fn(&tx{s: s}); err != nil {
		s.restore(snap)
		return err
	}
	return nil
}

type snapshot struct {
	accounts map[string]app.Account
	balances map[string]domain.Balance
	books    map[string]domain.Book
	orders   map[string]domain.Order
	trades   []app.StoredTrade
}

func (s *Store) snapshot() snapshot {
	sn := snapshot{
		accounts: make(map[string]app.Account, len(s.accounts)),
		balances: make(map[string]domain.Balance, len(s.balances)),
		books:    make(map[string]domain.Book, len(s.books)),
		orders:   make(map[string]domain.Order, len(s.orders)),
		trades:   make([]app.StoredTrade, len(s.trades)),
	}
	for k, v := range s.accounts {
		sn.accounts[k] = v
	}
	for k, v := range s.balances {
		sn.balances[k] = v.Clone()
	}
	for k, v := range s.books {
		sn.books[k] = v.Clone()
	}
	for k, v := range s.orders {
		sn.orders[k] = v.Clone()
	}
	for i, tr := range s.trades {
		sn.trades[i] = app.StoredTrade{
			Instrument: tr.Instrument,
			CreatedAt:  tr.CreatedAt,
			Trade:      tr.Trade.Clone(),
		}
	}
	return sn
}

func (s *Store) restore(sn snapshot) {
	s.accounts = sn.accounts
	s.balances = make(map[string]*domain.Balance, len(sn.balances))
	for k, v := range sn.balances {
		cp := v
		s.balances[k] = &cp
	}
	s.books = make(map[string]*domain.Book, len(sn.books))
	for k, v := range sn.books {
		cp := v
		s.books[k] = &cp
	}
	s.orders = sn.orders
	s.trades = sn.trades
}

type tx struct {
	s *Store
}

func balanceKey(accountID string, asset domain.Asset) string {
	return accountID + "|" + string(asset)
}

func (t *tx) CreateAccount(label string, now time.Time, id string) app.Account {
	acc := app.Account{ID: id, Label: label, CreatedAt: now}
	t.s.accounts[id] = acc
	return acc
}

func (t *tx) Accounts() []app.Account {
	out := make([]app.Account, 0, len(t.s.accounts))
	for _, acc := range t.s.accounts {
		out = append(out, acc)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func (t *tx) Account(id string) (app.Account, error) {
	acc, ok := t.s.accounts[id]
	if !ok {
		return app.Account{}, app.ErrAccountNotFound
	}
	return acc, nil
}

func (t *tx) Balance(accountID string, asset domain.Asset) *domain.Balance {
	key := balanceKey(accountID, asset)
	bal, ok := t.s.balances[key]
	if !ok {
		created := domain.NewBalance(accountID, asset)
		bal = &created
		t.s.balances[key] = bal
	}
	return bal
}

func (t *tx) Book(instrument string) *domain.Book {
	book, ok := t.s.books[instrument]
	if !ok {
		created := domain.NewBook(instrument)
		book = &created
		t.s.books[instrument] = book
	}
	return book
}

func (t *tx) SaveOrder(o domain.Order) {
	t.s.orders[o.ID] = o.Clone()
}

func (t *tx) Order(id string) (domain.Order, error) {
	o, ok := t.s.orders[id]
	if !ok {
		return domain.Order{}, app.ErrOrderNotFound
	}
	return o.Clone(), nil
}

func (t *tx) OrdersByAccount(accountID string) []domain.Order {
	var out []domain.Order
	for _, o := range t.s.orders {
		if o.AccountID == accountID {
			out = append(out, o.Clone())
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		}
		return out[i].ID > out[j].ID
	})
	return out
}

func (t *tx) AddTrade(instrument string, tr domain.Trade, at time.Time) {
	t.s.trades = append(t.s.trades, app.StoredTrade{
		Instrument: instrument,
		CreatedAt:  at,
		Trade:      tr.Clone(),
	})
}

func (t *tx) Trades(instrument string) []app.StoredTrade {
	var out []app.StoredTrade
	for i := len(t.s.trades) - 1; i >= 0; i-- {
		tr := t.s.trades[i]
		if tr.Instrument != instrument {
			continue
		}
		out = append(out, app.StoredTrade{
			Instrument: tr.Instrument,
			CreatedAt:  tr.CreatedAt,
			Trade:      tr.Trade.Clone(),
		})
	}
	return out
}
