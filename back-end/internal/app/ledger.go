package app

import (
	"context"

	"github.com/srochagomes/mb-avaliacao/internal/domain"
)

type BalanceView struct {
	Asset     domain.Asset
	Available domain.Amount
	Reserved  domain.Amount
}

func (s *Service) Credit(ctx context.Context, accountID, asset string, amount domain.Amount) error {
	return s.move(ctx, accountID, asset, amount, true)
}

func (s *Service) Debit(ctx context.Context, accountID, asset string, amount domain.Amount) error {
	return s.move(ctx, accountID, asset, amount, false)
}

func (s *Service) move(ctx context.Context, accountID, asset string, amount domain.Amount, credit bool) error {
	parsed, err := parseAsset(asset)
	if err != nil {
		return err
	}
	return s.store.Within(ctx, func(tx Tx) error {
		if err := s.requireAccount(tx, accountID); err != nil {
			return err
		}
		bal := tx.Balance(accountID, parsed)
		if credit {
			return bal.Credit(amount)
		}
		return bal.Debit(amount)
	})
}

func (s *Service) Balances(ctx context.Context, accountID string) ([]BalanceView, error) {
	var views []BalanceView
	err := s.store.Within(ctx, func(tx Tx) error {
		if err := s.requireAccount(tx, accountID); err != nil {
			return err
		}
		for _, asset := range []domain.Asset{domain.AssetBRL, domain.AssetBTC} {
			bal := tx.Balance(accountID, asset)
			views = append(views, BalanceView{
				Asset:     asset,
				Available: bal.Available(),
				Reserved:  bal.Reserved(),
			})
		}
		return nil
	})
	return views, err
}
