package app

import (
	"context"

	"github.com/srochagomes/mb-avaliacao/internal/domain"
)

type BookView struct {
	Instrument string
	Bids       []domain.Order
	Asks       []domain.Order
}

func (s *Service) Book(ctx context.Context, instrument string) (BookView, error) {
	if instrument != Instrument {
		return BookView{}, ValidationError{Message: "invalid instrument"}
	}
	view := BookView{Instrument: instrument}
	err := s.store.Within(ctx, func(tx Tx) error {
		book := tx.Book(instrument)
		view.Bids = book.Bids()
		view.Asks = book.Asks()
		return nil
	})
	return view, err
}

func (s *Service) Trades(ctx context.Context, instrument string) ([]StoredTrade, error) {
	if instrument != Instrument {
		return nil, ValidationError{Message: "invalid instrument"}
	}
	var trades []StoredTrade
	err := s.store.Within(ctx, func(tx Tx) error {
		trades = tx.Trades(instrument)
		return nil
	})
	return trades, err
}
