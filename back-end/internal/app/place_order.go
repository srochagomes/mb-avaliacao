package app

import (
	"context"

	"github.com/srochagomes/mb-avaliacao/internal/domain"
)

type PlaceOrderInput struct {
	AccountID  string
	Instrument string
	Side       domain.Side
	Price      domain.Amount
	Quantity   domain.Amount
}

type PlaceOrderResult struct {
	Order  domain.Order
	Trades []domain.Trade
}

func (s *Service) PlaceOrder(ctx context.Context, in PlaceOrderInput) (PlaceOrderResult, error) {
	if in.Instrument != Instrument {
		return PlaceOrderResult{}, ValidationError{Message: "invalid instrument"}
	}
	asset, err := assetOf(in.Side)
	if err != nil {
		return PlaceOrderResult{}, err
	}
	if in.Price.Cmp(domain.Amount{}) <= 0 || in.Quantity.Cmp(domain.Amount{}) <= 0 {
		return PlaceOrderResult{}, domain.ErrInvalidAmount
	}

	var result PlaceOrderResult
	err = s.store.Within(ctx, func(tx Tx) error {
		if err := s.requireAccount(tx, in.AccountID); err != nil {
			return err
		}
		order := domain.Order{
			ID:                s.newID(),
			AccountID:         in.AccountID,
			Instrument:        in.Instrument,
			Side:              in.Side,
			Price:             in.Price,
			OriginalQuantity:  in.Quantity,
			RemainingQuantity: in.Quantity,
			Status:            domain.StatusOpen,
			CreatedAt:         s.now().UTC(),
		}
		if err := domain.ReserveForOrder(tx.Balance(order.AccountID, asset), order); err != nil {
			return err
		}
		book := tx.Book(order.Instrument)
		trades, final, err := book.Match(order)
		if err != nil {
			return err
		}
		for _, tr := range trades {
			limit := tr.Price
			if order.Side == domain.SideBuy {
				limit = order.Price
			}
			if err := domain.Settle(
				tx.Balance(tr.BuyerAccountID, domain.AssetBRL),
				tx.Balance(tr.BuyerAccountID, domain.AssetBTC),
				tx.Balance(tr.SellerAccountID, domain.AssetBRL),
				tx.Balance(tr.SellerAccountID, domain.AssetBTC),
				tr,
				limit,
			); err != nil {
				return err
			}
			if err := fillMaker(tx, tr); err != nil {
				return err
			}
			tx.AddTrade(order.Instrument, tr, order.CreatedAt)
		}
		tx.SaveOrder(final)
		result.Order = final.Clone()
		result.Trades = append([]domain.Trade(nil), trades...)
		return nil
	})
	return result, err
}

func fillMaker(tx Tx, tr domain.Trade) error {
	maker, err := tx.Order(tr.MakerOrderID)
	if err != nil {
		return err
	}
	next, err := maker.RemainingQuantity.Sub(tr.Quantity)
	if err != nil {
		return err
	}
	maker.RemainingQuantity = next
	if next.IsZero() {
		maker.Status = domain.StatusFilled
	}
	tx.SaveOrder(maker)
	return nil
}
