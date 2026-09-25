package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/srochagomes/mb-avaliacao/internal/app"
	"github.com/srochagomes/mb-avaliacao/internal/memory"
)

func TestEnunciado(t *testing.T) {
	api := newAPI(t)
	a := api.createAccount("Conta A")
	b := api.createAccount("Conta B")
	api.credit(a, "BRL", "500000")
	api.credit(b, "BTC", "1")

	buy := api.place(a, "buy", "500000", "1")
	if buy.Order.Status != "open" || len(buy.Trades) != 0 {
		t.Fatalf("buy resting: %+v", buy)
	}
	sell := api.place(b, "sell", "500000", "1")
	if sell.Order.Status != "filled" || len(sell.Trades) != 1 || sell.Trades[0].Price != "500000" {
		t.Fatalf("sell: %+v", sell)
	}

	abrl, abtc := api.balance(a)
	bbrl, bbtc := api.balance(b)
	if abrl != "0" || abtc != "1" || bbrl != "500000" || bbtc != "0" {
		t.Fatalf("A brl %s btc %s; B brl %s btc %s", abrl, abtc, bbrl, bbtc)
	}
	book := api.book()
	if len(book.Bids) != 0 || len(book.Asks) != 0 {
		t.Fatalf("book %+v", book)
	}
}

func TestSurplusAndCancel(t *testing.T) {
	api := newAPI(t)
	a := api.createAccount("A")
	b := api.createAccount("B")
	api.credit(b, "BTC", "1")
	api.credit(a, "BRL", "510000")
	api.place(b, "sell", "500000", "1")
	got := api.place(a, "buy", "510000", "1")
	if len(got.Trades) != 1 || got.Trades[0].Price != "500000" {
		t.Fatalf("trade %+v", got)
	}
	abrl, abtc := api.balance(a)
	if abrl != "10000" || abtc != "1" {
		t.Fatalf("A brl %s btc %s", abrl, abtc)
	}

	api.credit(a, "BRL", "40")
	resting := api.place(a, "buy", "40", "1")
	if resting.Order.Status != "open" {
		t.Fatalf("status %s", resting.Order.Status)
	}
	canceled := api.cancel(resting.Order.ID)
	if canceled.Status != "canceled" {
		t.Fatalf("cancel %+v", canceled)
	}
	abrl, _ = api.balance(a)
	if abrl != "10040" {
		t.Fatalf("after cancel brl %s", abrl)
	}
	if status, _ := api.cancelStatus(resting.Order.ID); status != http.StatusUnprocessableEntity {
		t.Fatalf("second cancel %d", status)
	}
}

func TestInsufficientBalanceDoesNotEnterBook(t *testing.T) {
	api := newAPI(t)
	a := api.createAccount("A")
	api.credit(a, "BRL", "10")
	status := api.placeStatus(a, "buy", "40", "1")
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d", status)
	}
	brl, _ := api.balance(a)
	if brl != "10" {
		t.Fatalf("brl %s", brl)
	}
	book := api.book()
	if len(book.Bids) != 0 {
		t.Fatalf("bids %d", len(book.Bids))
	}
}

func TestDebitIgnoresReserved(t *testing.T) {
	api := newAPI(t)
	a := api.createAccount("A")
	api.credit(a, "BRL", "100")
	api.place(a, "buy", "40", "1")
	if status := api.debitStatus(a, "BRL", "70"); status != http.StatusUnprocessableEntity {
		t.Fatalf("debit reserved %d", status)
	}
	api.debit(a, "BRL", "60")
	views := api.balances(a)
	if views["BRL"].Available != "0" || views["BRL"].Reserved != "40" {
		t.Fatalf("%+v", views["BRL"])
	}
}

type apiClient struct {
	t   *testing.T
	srv *httptest.Server
}

func newAPI(t *testing.T) *apiClient {
	t.Helper()
	srv := httptest.NewServer(Handler(app.New(memory.New())))
	t.Cleanup(srv.Close)
	return &apiClient{t: t, srv: srv}
}

func (c *apiClient) createAccount(label string) string {
	c.t.Helper()
	var out struct {
		ID string `json:"id"`
	}
	c.do(http.MethodPost, "/v1/accounts", map[string]string{"label": label}, http.StatusCreated, &out)
	return out.ID
}

func (c *apiClient) credit(id, asset, amount string) {
	c.t.Helper()
	c.do(http.MethodPost, "/v1/accounts/"+id+"/credits", map[string]string{"asset": asset, "amount": amount}, http.StatusOK, nil)
}

func (c *apiClient) debit(id, asset, amount string) {
	c.t.Helper()
	c.do(http.MethodPost, "/v1/accounts/"+id+"/debits", map[string]string{"asset": asset, "amount": amount}, http.StatusOK, nil)
}

func (c *apiClient) debitStatus(id, asset, amount string) int {
	c.t.Helper()
	return c.status(http.MethodPost, "/v1/accounts/"+id+"/debits", map[string]string{"asset": asset, "amount": amount})
}

type placed struct {
	Order struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"order"`
	Trades []struct {
		Price string `json:"price"`
	} `json:"trades"`
}

func (c *apiClient) place(id, side, price, qty string) placed {
	c.t.Helper()
	var out placed
	c.do(http.MethodPost, "/v1/orders", map[string]string{
		"account_id": id, "instrument": "BTC-BRL", "side": side, "price": price, "quantity": qty,
	}, http.StatusCreated, &out)
	return out
}

func (c *apiClient) placeStatus(id, side, price, qty string) int {
	c.t.Helper()
	return c.status(http.MethodPost, "/v1/orders", map[string]string{
		"account_id": id, "instrument": "BTC-BRL", "side": side, "price": price, "quantity": qty,
	})
}

func (c *apiClient) cancel(id string) orderJSON {
	c.t.Helper()
	var out orderJSON
	c.do(http.MethodDelete, "/v1/orders/"+id, nil, http.StatusOK, &out)
	return out
}

func (c *apiClient) cancelStatus(id string) (int, orderJSON) {
	c.t.Helper()
	status, raw := c.raw(http.MethodDelete, "/v1/orders/"+id, nil)
	var out orderJSON
	_ = json.Unmarshal(raw, &out)
	return status, out
}

func (c *apiClient) balance(id string) (brl, btc string) {
	c.t.Helper()
	views := c.balances(id)
	return views["BRL"].Available, views["BTC"].Available
}

func (c *apiClient) balances(id string) map[string]balanceJSON {
	c.t.Helper()
	var out []balanceJSON
	c.do(http.MethodGet, "/v1/accounts/"+id+"/balances", nil, http.StatusOK, &out)
	views := map[string]balanceJSON{}
	for _, v := range out {
		views[v.Asset] = v
	}
	return views
}

type bookBody struct {
	Bids []orderJSON `json:"bids"`
	Asks []orderJSON `json:"asks"`
}

func (c *apiClient) book() bookBody {
	c.t.Helper()
	var out bookBody
	c.do(http.MethodGet, "/v1/books/BTC-BRL", nil, http.StatusOK, &out)
	return out
}

func (c *apiClient) do(method, path string, body any, want int, dst any) {
	c.t.Helper()
	status, raw := c.raw(method, path, body)
	if status != want {
		c.t.Fatalf("%s %s: status %d body %s", method, path, status, raw)
	}
	if dst != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, dst); err != nil {
			c.t.Fatal(err)
		}
	}
}

func (c *apiClient) status(method, path string, body any) int {
	c.t.Helper()
	status, _ := c.raw(method, path, body)
	return status
}

func (c *apiClient) raw(method, path string, body any) (int, []byte) {
	c.t.Helper()
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, c.srv.URL+path, reader)
	if err != nil {
		c.t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		c.t.Fatal(err)
	}
	return res.StatusCode, raw
}
