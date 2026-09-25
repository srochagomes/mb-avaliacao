package domain

func ReserveForOrder(bal *Balance, o Order) error {
    if o.Side == SideBuy {
        if bal.Asset != AssetBRL {
            return ErrInvalidAmount
        }
        return bal.Reserve(o.Price.MulDiv(o.RemainingQuantity))
    }
    if o.Side == SideSell {
        if bal.Asset != AssetBTC {
            return ErrInvalidAmount
        }
        return bal.Reserve(o.RemainingQuantity)
    }
    return ErrInvalidAmount
}

func Settle(buyerBRL, buyerBTC, sellerBRL, sellerBTC *Balance, tr Trade, buyerLimit Amount) error {
    notional := tr.Price.MulDiv(tr.Quantity)
    if err := sellerBTC.ConsumeReserved(tr.Quantity); err != nil {
        return err
    }
    if err := sellerBRL.Credit(notional); err != nil {
        return err
    }
    if err := buyerBRL.ConsumeReserved(notional); err != nil {
        return err
    }
    diff, err := buyerLimit.Sub(tr.Price)
    if err != nil {
        return err
    }
    if !diff.IsZero() {
        if err := buyerBRL.Release(diff.MulDiv(tr.Quantity)); err != nil {
            return err
        }
    }
    return buyerBTC.Credit(tr.Quantity)
}

func ReleaseForOrder(bal *Balance, o Order) error {
    if !o.IsOpen() {
        return ErrOrderNotOpen
    }
    if o.Side == SideBuy {
        if bal.Asset != AssetBRL {
            return ErrInvalidAmount
        }
        return bal.Release(o.Price.MulDiv(o.RemainingQuantity))
    }
    if o.Side == SideSell {
        if bal.Asset != AssetBTC {
            return ErrInvalidAmount
        }
        return bal.Release(o.RemainingQuantity)
    }
    return ErrInvalidAmount
}