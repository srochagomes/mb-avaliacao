package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/srochagomes/mb-avaliacao/internal/app"
	"github.com/srochagomes/mb-avaliacao/internal/domain"
)

func Handler(svc *app.Service) http.Handler {
	mux := http.NewServeMux()
	h := &handler{svc: svc}
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /v1/accounts", h.createAccount)
	mux.HandleFunc("GET /v1/accounts", h.listAccounts)
	mux.HandleFunc("POST /v1/accounts/{id}/credits", h.credit)
	mux.HandleFunc("POST /v1/accounts/{id}/debits", h.debit)
	mux.HandleFunc("GET /v1/accounts/{id}/balances", h.balances)
	mux.HandleFunc("GET /v1/accounts/{id}/orders", h.orders)
	mux.HandleFunc("POST /v1/orders", h.placeOrder)
	mux.HandleFunc("DELETE /v1/orders/{id}", h.cancelOrder)
	mux.HandleFunc("GET /v1/books/{instrument}", h.book)
	mux.HandleFunc("GET /v1/trades/{instrument}", h.trades)
	return mux
}

type handler struct {
	svc *app.Service
}

type errBody struct {
	Error string `json:"error"`
}

func (h *handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) createAccount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label string `json:"label"`
	}
	if err := decode(w, r, &body); err != nil {
		writeErr(w, err)
		return
	}
	acc, err := h.svc.CreateAccount(r.Context(), body.Label)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAccountJSON(acc))
}

func (h *handler) listAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.svc.ListAccounts(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]any, 0, len(accounts))
	for _, acc := range accounts {
		out = append(out, toAccountJSON(acc))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *handler) credit(w http.ResponseWriter, r *http.Request) {
	h.move(w, r, true)
}

func (h *handler) debit(w http.ResponseWriter, r *http.Request) {
	h.move(w, r, false)
}

func (h *handler) move(w http.ResponseWriter, r *http.Request, credit bool) {
	var body struct {
		Asset  string `json:"asset"`
		Amount string `json:"amount"`
	}
	if err := decode(w, r, &body); err != nil {
		writeErr(w, err)
		return
	}
	amount, err := domain.ParseAmount(body.Amount)
	if err != nil {
		writeErr(w, app.ValidationError{Message: "invalid amount"})
		return
	}
	id := r.PathValue("id")
	if credit {
		err = h.svc.Credit(r.Context(), id, body.Asset, amount)
	} else {
		err = h.svc.Debit(r.Context(), id, body.Asset, amount)
	}
	if err != nil {
		writeErr(w, err)
		return
	}
	h.writeBalances(w, r, id)
}

func (h *handler) balances(w http.ResponseWriter, r *http.Request) {
	h.writeBalances(w, r, r.PathValue("id"))
}

func (h *handler) writeBalances(w http.ResponseWriter, r *http.Request, id string) {
	views, err := h.svc.Balances(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]balanceJSON, 0, len(views))
	for _, v := range views {
		out = append(out, balanceJSON{
			Asset:     string(v.Asset),
			Available: v.Available.String(),
			Reserved:  v.Reserved.String(),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *handler) orders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.svc.Orders(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ordersJSON(orders))
}

func (h *handler) placeOrder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccountID  string `json:"account_id"`
		Instrument string `json:"instrument"`
		Side       string `json:"side"`
		Price      string `json:"price"`
		Quantity   string `json:"quantity"`
	}
	if err := decode(w, r, &body); err != nil {
		writeErr(w, err)
		return
	}
	price, err := domain.ParseAmount(body.Price)
	if err != nil {
		writeErr(w, app.ValidationError{Message: "invalid amount"})
		return
	}
	qty, err := domain.ParseAmount(body.Quantity)
	if err != nil {
		writeErr(w, app.ValidationError{Message: "invalid amount"})
		return
	}
	result, err := h.svc.PlaceOrder(r.Context(), app.PlaceOrderInput{
		AccountID:  body.AccountID,
		Instrument: body.Instrument,
		Side:       domain.Side(body.Side),
		Price:      price,
		Quantity:   qty,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"order":  toOrderJSON(result.Order),
		"trades": tradesJSON(result.Trades),
	})
}

func (h *handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	order, err := h.svc.CancelOrder(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toOrderJSON(order))
}

func (h *handler) book(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.Book(r.Context(), r.PathValue("instrument"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"instrument": view.Instrument,
		"bids":       ordersJSON(view.Bids),
		"asks":       ordersJSON(view.Asks),
	})
}

func (h *handler) trades(w http.ResponseWriter, r *http.Request) {
	trades, err := h.svc.Trades(r.Context(), r.PathValue("instrument"))
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]tradeJSON, 0, len(trades))
	for _, tr := range trades {
		item := tradeJSONFrom(tr.Trade)
		item.CreatedAt = tr.CreatedAt
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, out)
}

func decode(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return app.ValidationError{Message: "invalid json"}
	}
	return nil
}

func writeErr(w http.ResponseWriter, err error) {
	var validation app.ValidationError
	switch {
	case errors.As(err, &validation):
		writeJSON(w, http.StatusBadRequest, errBody{Error: validation.Error()})
	case errors.Is(err, app.ErrAccountNotFound), errors.Is(err, app.ErrOrderNotFound):
		writeJSON(w, http.StatusNotFound, errBody{Error: err.Error()})
	case errors.Is(err, domain.ErrInsufficientBalance),
		errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrOrderNotOpen),
		errors.Is(err, domain.ErrInstrumentMismatch):
		writeJSON(w, http.StatusUnprocessableEntity, errBody{Error: err.Error()})
	default:
		log.Printf("request failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, errBody{Error: "internal error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode: %v", err)
	}
}

type accountJSON struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"created_at"`
}

func toAccountJSON(acc app.Account) accountJSON {
	return accountJSON{ID: acc.ID, Label: acc.Label, CreatedAt: acc.CreatedAt}
}

type balanceJSON struct {
	Asset     string `json:"asset"`
	Available string `json:"available"`
	Reserved  string `json:"reserved"`
}

type orderJSON struct {
	ID                string    `json:"id"`
	AccountID         string    `json:"account_id"`
	Instrument        string    `json:"instrument"`
	Side              string    `json:"side"`
	Price             string    `json:"price"`
	Quantity          string    `json:"quantity"`
	RemainingQuantity string    `json:"remaining_quantity"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

func toOrderJSON(o domain.Order) orderJSON {
	return orderJSON{
		ID:                o.ID,
		AccountID:         o.AccountID,
		Instrument:        o.Instrument,
		Side:              string(o.Side),
		Price:             o.Price.String(),
		Quantity:          o.OriginalQuantity.String(),
		RemainingQuantity: o.RemainingQuantity.String(),
		Status:            string(o.Status),
		CreatedAt:         o.CreatedAt,
	}
}

func ordersJSON(orders []domain.Order) []orderJSON {
	out := make([]orderJSON, 0, len(orders))
	for _, o := range orders {
		out = append(out, toOrderJSON(o))
	}
	return out
}

type tradeJSON struct {
	Price           string    `json:"price"`
	Quantity        string    `json:"quantity"`
	MakerOrderID    string    `json:"maker_order_id"`
	TakerOrderID    string    `json:"taker_order_id"`
	BuyerAccountID  string    `json:"buyer_account_id"`
	SellerAccountID string    `json:"seller_account_id"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
}

func tradeJSONFrom(tr domain.Trade) tradeJSON {
	return tradeJSON{
		Price:           tr.Price.String(),
		Quantity:        tr.Quantity.String(),
		MakerOrderID:    tr.MakerOrderID,
		TakerOrderID:    tr.TakerOrderID,
		BuyerAccountID:  tr.BuyerAccountID,
		SellerAccountID: tr.SellerAccountID,
	}
}

func tradesJSON(trades []domain.Trade) []tradeJSON {
	out := make([]tradeJSON, 0, len(trades))
	for _, tr := range trades {
		out = append(out, tradeJSONFrom(tr))
	}
	return out
}
