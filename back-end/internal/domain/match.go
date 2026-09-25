package domain

import "slices"

type Trade struct {
    Price            Amount // preço do maker
    Quantity         Amount
    MakerOrderID     string
    TakerOrderID     string
    BuyerAccountID   string
    SellerAccountID  string
}

func (b *Book) Match(taker Order) ([]Trade, Order, error) {
    if taker.Instrument != b.Instrument {
        return nil, taker, ErrInstrumentMismatch
    }
    if !taker.IsOpen() {
        return nil, taker, ErrOrderNotOpen
    }
    resting := &b.asks
    if taker.Side == SideSell {
        resting = &b.bids
    }
    var trades []Trade
    for i := 0; i < len(*resting) && !taker.RemainingQuantity.IsZero(); {
        maker := (*resting)[i]
        if !crosses(taker, maker) {
            break
        }
        if maker.AccountID == taker.AccountID {
            i++
            continue
        }
        qty := taker.RemainingQuantity
        if maker.RemainingQuantity.Cmp(qty) < 0 {
            qty = maker.RemainingQuantity
        }
        trades = append(trades, trade(taker, maker, qty))
        var err error
        taker.RemainingQuantity, err = taker.RemainingQuantity.Sub(qty)
        if err != nil {
            return nil, taker, err
        }
        maker.RemainingQuantity, err = maker.RemainingQuantity.Sub(qty)
        if err != nil {
            return nil, taker, err
        }
        if maker.RemainingQuantity.IsZero() {
            *resting = slices.Delete(*resting, i, i+1)
            continue
        }
        (*resting)[i] = maker
        i++
    }
    if taker.RemainingQuantity.IsZero() {
        taker.Status = StatusFilled
        return trades, taker, nil
    }
    taker.Status = StatusOpen
    if err := b.Add(taker); err != nil {
        return nil, taker, err
    }
    return trades, taker, nil
}


func crosses(taker, maker Order) bool {
    if taker.Side == SideBuy {
        return maker.Price.Cmp(taker.Price) <= 0
    }
    return maker.Price.Cmp(taker.Price) >= 0
}

func trade(taker, maker Order, qty Amount) Trade {
    tr := Trade{
        Price:        maker.Price,
        Quantity:     qty,
        MakerOrderID: maker.ID,
        TakerOrderID: taker.ID,
    }
    if taker.Side == SideBuy {
        tr.BuyerAccountID = taker.AccountID
        tr.SellerAccountID = maker.AccountID
        return tr
    }
    tr.BuyerAccountID = maker.AccountID
    tr.SellerAccountID = taker.AccountID
    return tr
}