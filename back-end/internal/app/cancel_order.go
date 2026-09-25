package app

import (
	"context"

	"github.com/srochagomes/mb-avaliacao/internal/domain"
)

func (s *Service) CancelOrder(ctx context.Context, id string) (domain.Order, error) {
	if id == "" {
		return domain.Order{}, ValidationError{Message: "order id required"}
	}
	var canceled domain.Order
	err := s.store.Within(ctx, func(tx Tx) error {
		order, err := tx.Order(id)
		if err != nil {
			return err
		}
		if !order.IsOpen() {
			return domain.ErrOrderNotOpen
		}
		asset, err := assetOf(order.Side)
		if err != nil {
			return err
		}
		removed, err := tx.Book(order.Instrument).Remove(id)
		if err != nil {
			return err
		}
		if err := domain.ReleaseForOrder(tx.Balance(order.AccountID, asset), removed); err != nil {
			return err
		}
		removed.Status = domain.StatusCanceled
		tx.SaveOrder(removed)
		canceled = removed.Clone()
		return nil
	})
	return canceled, err
}

func (s *Service) Orders(ctx context.Context, accountID string) ([]domain.Order, error) {
	var orders []domain.Order
	err := s.store.Within(ctx, func(tx Tx) error {
		if err := s.requireAccount(tx, accountID); err != nil {
			return err
		}
		orders = tx.OrdersByAccount(accountID)
		return nil
	})
	return orders, err
}
