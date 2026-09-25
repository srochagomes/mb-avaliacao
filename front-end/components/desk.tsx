"use client";

import { FormEvent, ReactNode, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ArrowDownToLine, ArrowUpFromLine, BookOpen, Plus, Scale, Wallet } from "lucide-react";
import { Account, ApiError, Balance, Book, Order, Trade, api } from "@/lib/api";
import { formatAmount, shortId } from "@/lib/format";

type Notice = { tone: "ok" | "err"; text: string };
type Tab = "saldo" | "ordens" | "mercado";

const statusLabel: Record<Order["status"], string> = {
  open: "Aberta",
  filled: "Executada",
  canceled: "Cancelada",
};

const tabs: { id: Tab; label: string; hint: string; icon: ReactNode }[] = [
  { id: "saldo", label: "Carteiras e saldo", hint: "Criar conta, creditar e debitar", icon: <Wallet size={16} /> },
  { id: "ordens", label: "Ordens", hint: "Compra, venda e cancelamento", icon: <Scale size={16} /> },
  { id: "mercado", label: "Livro e negócios", hint: "Ofertas abertas e histórico", icon: <BookOpen size={16} /> },
];

export function Desk() {
  const [tab, setTab] = useState<Tab>("saldo");
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [selectedId, setSelectedId] = useState<string>("");
  const [balances, setBalances] = useState<Balance[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [book, setBook] = useState<Book>({ instrument: "BTC-BRL", bids: [], asks: [] });
  const [trades, setTrades] = useState<Trade[]>([]);
  const [notice, setNotice] = useState<Notice | null>(null);
  const [pending, setPending] = useState(false);
  const [label, setLabel] = useState("");
  const [creditAsset, setCreditAsset] = useState<"BRL" | "BTC">("BRL");
  const [creditAmount, setCreditAmount] = useState("");
  const [debitAsset, setDebitAsset] = useState<"BRL" | "BTC">("BRL");
  const [debitAmount, setDebitAmount] = useState("");
  const [side, setSide] = useState<"buy" | "sell">("buy");
  const [price, setPrice] = useState("500000");
  const [quantity, setQuantity] = useState("1");

  const selected = accounts.find((account) => account.id === selectedId);
  const selectedRef = useRef(selectedId);
  selectedRef.current = selectedId;
  const names = useMemo(() => new Map(accounts.map((account) => [account.id, account.label])), [accounts]);
  const lastTrade = trades[0];

  const fail = (error: unknown) => {
    setNotice({ tone: "err", text: error instanceof ApiError ? error.message : "Não foi possível concluir a operação." });
  };

  const loadMarket = useCallback(async () => {
    const [nextBook, nextTrades] = await Promise.all([
      api<Book>("/v1/books/BTC-BRL"),
      api<Trade[]>("/v1/trades/BTC-BRL"),
    ]);
    setBook(nextBook);
    setTrades(nextTrades);
  }, []);

  const loadWallet = useCallback(async (id: string) => {
    if (!id) return { balances: [] as Balance[], orders: [] as Order[] };
    const [balances, orders] = await Promise.all([
      api<Balance[]>(`/v1/accounts/${id}/balances`),
      api<Order[]>(`/v1/accounts/${id}/orders`),
    ]);
    return { balances, orders };
  }, []);

  const refresh = useCallback(async () => {
    const nextAccounts = await api<Account[]>("/v1/accounts");
    setAccounts(nextAccounts);
    setSelectedId((current) => current || nextAccounts[0]?.id || "");
    await loadMarket();
  }, [loadMarket]);

  useEffect(() => {
    refresh().catch(fail);
  }, [refresh]);

  useEffect(() => {
    let active = true;
    setBalances([]);
    setOrders([]);
    loadWallet(selectedId)
      .then((wallet) => {
        if (!active) return;
        setBalances(wallet.balances);
        setOrders(wallet.orders);
      })
      .catch((error) => {
        if (active) fail(error);
      });
    return () => {
      active = false;
    };
  }, [loadWallet, selectedId]);

  useEffect(() => {
    const timer = setInterval(() => {
      loadMarket().catch(() => undefined);
      if (selectedId) {
        const id = selectedId;
        loadWallet(id)
          .then((wallet) => {
            if (selectedRef.current !== id) return;
            setBalances(wallet.balances);
            setOrders(wallet.orders);
          })
          .catch(() => undefined);
      }
    }, 3000);
    return () => clearInterval(timer);
  }, [loadMarket, loadWallet, selectedId]);

  async function run(action: () => Promise<void>) {
    setPending(true);
    try {
      await action();
      await loadMarket();
      const wallet = await loadWallet(selectedId);
      setBalances(wallet.balances);
      setOrders(wallet.orders);
    } catch (error) {
      fail(error);
    } finally {
      setPending(false);
    }
  }

  function createWallet(event: FormEvent) {
    event.preventDefault();
    void run(async () => {
      const created = await api<Account>("/v1/accounts", {
        method: "POST",
        body: JSON.stringify({ label }),
      });
      setLabel("");
      setSelectedId(created.id);
      const nextAccounts = await api<Account[]>("/v1/accounts");
      setAccounts(nextAccounts);
      setNotice({ tone: "ok", text: `Carteira ${created.label} criada.` });
    });
  }

  function move(kind: "credits" | "debits", asset: string, amount: string, clear: () => void) {
    if (!selectedId) {
      setNotice({ tone: "err", text: "Crie e selecione uma carteira antes de movimentar saldo." });
      return;
    }
    void run(async () => {
      await api(`/v1/accounts/${selectedId}/${kind}`, {
        method: "POST",
        body: JSON.stringify({ asset, amount }),
      });
      clear();
      setNotice({
        tone: "ok",
        text: kind === "credits" ? "Crédito aplicado no disponível." : "Débito aplicado só no disponível.",
      });
    });
  }

  function place(event: FormEvent) {
    event.preventDefault();
    if (!selectedId) {
      setNotice({ tone: "err", text: "Selecione a carteira que envia a ordem." });
      return;
    }
    void run(async () => {
      const result = await api<{ order: Order; trades: Trade[] }>("/v1/orders", {
        method: "POST",
        body: JSON.stringify({
          account_id: selectedId,
          instrument: "BTC-BRL",
          side,
          price,
          quantity,
        }),
      });
      if (result.trades.length === 0) {
        setNotice({
          tone: "ok",
          text: "Sem contraparte neste preço. A ordem ficou aberta e o saldo foi reservado.",
        });
        return;
      }
      const last = result.trades[result.trades.length - 1];
      const summary = result.trades
        .map((trade) => `${formatAmount(trade.quantity)} BTC a ${formatAmount(trade.price)} BRL`)
        .join("; ");
      setNotice({
        tone: "ok",
        text: `Negócio no preço de quem já estava no livro: ${summary}. Último preço: ${formatAmount(last.price)}.`,
      });
      setTab("mercado");
    });
  }

  function cancel(id: string) {
    void run(async () => {
      await api(`/v1/orders/${id}`, { method: "DELETE" });
      setNotice({ tone: "ok", text: "Ordem cancelada. A reserva do restante voltou para o disponível." });
    });
  }

  const asks = [...book.asks].reverse();
  const openOrders = orders.filter((order) => order.status === "open");

  return (
    <main className="mx-auto flex min-h-screen max-w-6xl flex-col gap-5 px-4 py-6 md:px-8">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="font-mono text-xs tracking-[0.22em] text-gold uppercase">Mercado Bitcoin</p>
          <h1 className="mt-1 text-3xl font-semibold tracking-tight">Livro BTC/BRL</h1>
          <p className="mt-1 max-w-xl text-sm text-white/60">
            Simulação em memória. Use as abas: primeiro saldo, depois ordens, depois o livro.
          </p>
        </div>
        <LastTradeField trade={lastTrade} names={names} />
      </header>

      {notice && (
        <p
          role="status"
          className={`rounded-2xl border px-4 py-3 text-sm ${
            notice.tone === "err" ? "border-ask/40 bg-ask/10 text-ask" : "border-bid/30 bg-bid/10 text-bid"
          }`}
        >
          {notice.text}
        </p>
      )}

      <nav aria-label="Áreas da simulação" className="grid gap-2 sm:grid-cols-3">
        {tabs.map((item) => {
          const active = tab === item.id;
          return (
            <button
              key={item.id}
              type="button"
              onClick={() => setTab(item.id)}
              className={`rounded-2xl border px-4 py-3 text-left transition ${
                active ? "border-gold/50 bg-gold/10" : "border-white/10 bg-panel hover:border-white/20"
              }`}
            >
              <span className="flex items-center gap-2 text-sm font-medium">
                {item.icon}
                {item.label}
              </span>
              <span className="mt-1 block text-xs text-white/45">{item.hint}</span>
            </button>
          );
        })}
      </nav>

      {tab === "saldo" && (
        <section className="grid gap-4 lg:grid-cols-[280px_1fr]">
          <Panel title="Carteiras">
            <form onSubmit={createWallet} className="flex gap-2">
              <input
                aria-label="Nome da carteira"
                value={label}
                onChange={(event) => setLabel(event.target.value)}
                placeholder="Conta A"
                className={fieldClass}
              />
              <button className={iconButton} disabled={pending} aria-label="Criar carteira">
                <Plus size={18} />
              </button>
            </form>
            <ul className="mt-3 flex flex-col gap-2">
              {accounts.length === 0 && <Empty>Nenhuma carteira ainda.</Empty>}
              {accounts.map((account) => (
                <li key={account.id}>
                  <button
                    type="button"
                    onClick={() => setSelectedId(account.id)}
                    className={`flex w-full items-center gap-3 rounded-xl border px-3 py-2 text-left ${
                      account.id === selectedId ? "border-gold/50 bg-gold/10" : "border-white/10 bg-white/5"
                    }`}
                  >
                    <Wallet size={16} className="text-gold" />
                    <span>
                      <span className="block text-sm">{account.label}</span>
                      <span className="font-mono text-[11px] text-white/40">{shortId(account.id)}</span>
                    </span>
                  </button>
                </li>
              ))}
            </ul>
            <details className="mt-4 rounded-xl border border-white/10 px-3 py-2 text-sm text-white/70">
              <summary className="cursor-pointer text-white">Roteiro do enunciado</summary>
              <ol className="mt-2 list-decimal space-y-1 pl-4 text-xs">
                <li>Nesta aba: crie Conta A e Conta B.</li>
                <li>Credite 500000 BRL em A e 1 BTC em B.</li>
                <li>Aba Ordens: A compra 1 a 500000.</li>
                <li>Aba Ordens: B vende 1 a 500000.</li>
              </ol>
            </details>
          </Panel>

          <div className="flex flex-col gap-4">
            <Panel title={selected ? `Saldo · ${selected.label}` : "Selecione uma carteira"}>
              <div className="grid gap-3 sm:grid-cols-2">
                {(["BRL", "BTC"] as const).map((asset) => {
                  const row = balances.find((item) => item.asset === asset);
                  return (
                    <div key={asset} className="rounded-xl border border-white/10 bg-black/20 p-4">
                      <p className="font-mono text-xs text-white/50">{asset}</p>
                      <p className="mt-1 font-mono text-2xl">{formatAmount(row?.available ?? "0")}</p>
                      <p className="text-xs text-white/45">disponível</p>
                      <p className="mt-3 font-mono text-sm text-gold">
                        {formatAmount(row?.reserved ?? "0")} reservado
                      </p>
                    </div>
                  );
                })}
              </div>
            </Panel>

            <Panel title="Movimentar saldo">
              <p className="mb-3 text-sm text-white/50">
                Creditar e debitar entram só no disponível. Saldo reservado em ordem aberta não sai por débito.
              </p>
              <div className="grid gap-3 sm:grid-cols-2">
                <MoveForm
                  title="Creditar"
                  icon={<ArrowDownToLine size={16} />}
                  asset={creditAsset}
                  amount={creditAmount}
                  onAsset={setCreditAsset}
                  onAmount={setCreditAmount}
                  disabled={pending}
                  onSubmit={(event) => {
                    event.preventDefault();
                    move("credits", creditAsset, creditAmount, () => setCreditAmount(""));
                  }}
                />
                <MoveForm
                  title="Debitar"
                  icon={<ArrowUpFromLine size={16} />}
                  asset={debitAsset}
                  amount={debitAmount}
                  onAsset={setDebitAsset}
                  onAmount={setDebitAmount}
                  disabled={pending}
                  onSubmit={(event) => {
                    event.preventDefault();
                    move("debits", debitAsset, debitAmount, () => setDebitAmount(""));
                  }}
                />
              </div>
            </Panel>
          </div>
        </section>
      )}

      {tab === "ordens" && (
        <section className="grid gap-4 lg:grid-cols-2">
          <Panel title="Nova ordem limitada">
            {!selected && <Empty>Selecione uma carteira na aba de saldo.</Empty>}
            {selected && (
              <>
                <p className="mb-3 text-sm text-white/55">
                  Enviando pela carteira <span className="text-gold">{selected.label}</span>. A compra reserva BRL; a
                  venda reserva BTC.
                </p>
                <form onSubmit={place} className="flex flex-col gap-3">
                  <div className="grid grid-cols-2 gap-2">
                    {(["buy", "sell"] as const).map((value) => (
                      <button
                        key={value}
                        type="button"
                        onClick={() => setSide(value)}
                        className={`rounded-xl border px-3 py-2 text-sm ${
                          side === value
                            ? value === "buy"
                              ? "border-bid/50 bg-bid/15 text-bid"
                              : "border-ask/50 bg-ask/15 text-ask"
                            : "border-white/10 text-white/60"
                        }`}
                      >
                        {value === "buy" ? "Compra" : "Venda"}
                      </button>
                    ))}
                  </div>
                  <label className="text-xs text-white/50">
                    Preço limite (BRL)
                    <input
                      className={`${fieldClass} mt-1`}
                      value={price}
                      onChange={(event) => setPrice(event.target.value)}
                    />
                  </label>
                  <label className="text-xs text-white/50">
                    Quantidade (BTC)
                    <input
                      className={`${fieldClass} mt-1`}
                      value={quantity}
                      onChange={(event) => setQuantity(event.target.value)}
                    />
                  </label>
                  <button className={primaryButton} disabled={pending}>
                    Colocar ordem
                  </button>
                </form>
              </>
            )}
          </Panel>

          <Panel title={`Ordens · ${selected?.label ?? "—"}`}>
            {orders.length === 0 && <Empty>Nenhuma ordem desta carteira.</Empty>}
            {openOrders.length > 0 && (
              <p className="mb-2 text-xs text-white/45">{openOrders.length} aberta(s) no livro</p>
            )}
            <ul className="flex flex-col gap-2">
              {orders.map((order) => (
                <li
                  key={order.id}
                  className="flex items-center justify-between gap-3 rounded-xl border border-white/10 px-3 py-2"
                >
                  <div>
                    <p className={`text-sm ${order.side === "buy" ? "text-bid" : "text-ask"}`}>
                      {order.side === "buy" ? "Compra" : "Venda"} {formatAmount(order.remaining_quantity)} /{" "}
                      {formatAmount(order.quantity)} BTC
                    </p>
                    <p className="font-mono text-xs text-white/45">
                      {formatAmount(order.price)} · {statusLabel[order.status]}
                    </p>
                  </div>
                  {order.status === "open" && (
                    <button
                      type="button"
                      className={ghostButton}
                      disabled={pending}
                      onClick={() => cancel(order.id)}
                    >
                      Cancelar
                    </button>
                  )}
                </li>
              ))}
            </ul>
          </Panel>
        </section>
      )}

      {tab === "mercado" && (
        <section className="grid gap-4 lg:grid-cols-2">
          <Panel title="Livro de ofertas">
            <LevelList rows={asks} tone="ask" names={names} selectedId={selectedId} empty="Sem ofertas de venda." />
            <div className="my-3 rounded-xl border border-dashed border-white/15 px-3 py-2 text-center">
              <p className="font-mono text-[10px] tracking-[0.2em] text-white/35 uppercase">Último preço</p>
              <p className="font-mono text-lg text-gold">
                {lastTrade ? formatAmount(lastTrade.price) : "—"}
              </p>
            </div>
            <LevelList rows={book.bids} tone="bid" names={names} selectedId={selectedId} empty="Sem ofertas de compra." />
          </Panel>

          <Panel title="Histórico de negócios">
            {trades.length === 0 && <Empty>Nenhum negócio ainda.</Empty>}
            <ul className="flex flex-col gap-2">
              {trades.map((trade, index) => (
                <li
                  key={`${trade.taker_order_id}-${trade.maker_order_id}-${trade.quantity}-${index}`}
                  className={`rounded-xl border px-3 py-2 font-mono text-xs ${
                    index === 0 ? "border-gold/40 bg-gold/10 text-gold" : "border-white/10 text-white/75"
                  }`}
                >
                  {formatAmount(trade.quantity)} BTC a {formatAmount(trade.price)} BRL
                  <span className="mt-1 block text-white/40">
                    {names.get(trade.buyer_account_id) ?? shortId(trade.buyer_account_id)} comprou de{" "}
                    {names.get(trade.seller_account_id) ?? shortId(trade.seller_account_id)}
                  </span>
                </li>
              ))}
            </ul>
          </Panel>
        </section>
      )}
    </main>
  );
}

function LastTradeField({ trade, names }: { trade?: Trade; names: Map<string, string> }) {
  return (
    <label className="min-w-[220px] rounded-2xl border border-gold/30 bg-panel px-4 py-3">
      <span className="block font-mono text-[10px] tracking-[0.18em] text-white/45 uppercase">
        Última negociação
      </span>
      <input
        readOnly
        aria-label="Valor da última negociação"
        className="mt-1 w-full bg-transparent font-mono text-2xl text-gold outline-none"
        value={trade ? `${formatAmount(trade.price)} BRL` : "Sem negócios"}
      />
      <span className="mt-1 block text-xs text-white/40">
        {trade
          ? `${formatAmount(trade.quantity)} BTC · ${names.get(trade.buyer_account_id) ?? shortId(trade.buyer_account_id)} ← ${names.get(trade.seller_account_id) ?? shortId(trade.seller_account_id)}`
          : "Atualiza quando uma ordem casa"}
      </span>
    </label>
  );
}

function Panel({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="rounded-2xl border border-white/10 bg-panel p-4 shadow-[0_20px_60px_rgba(0,0,0,0.25)]">
      <h2 className="mb-3 text-sm font-medium text-white/80">{title}</h2>
      {children}
    </section>
  );
}

function Empty({ children }: { children: ReactNode }) {
  return <p className="text-sm text-white/40">{children}</p>;
}

function MoveForm({
  title,
  icon,
  asset,
  amount,
  onAsset,
  onAmount,
  onSubmit,
  disabled,
}: {
  title: string;
  icon: ReactNode;
  asset: "BRL" | "BTC";
  amount: string;
  onAsset: (asset: "BRL" | "BTC") => void;
  onAmount: (amount: string) => void;
  onSubmit: (event: FormEvent) => void;
  disabled: boolean;
}) {
  return (
    <form onSubmit={onSubmit} className="rounded-xl border border-white/10 p-3">
      <p className="mb-2 flex items-center gap-2 text-sm text-white/80">
        {icon}
        {title}
      </p>
      <div className="flex gap-2">
        <select
          aria-label={`Ativo para ${title}`}
          className={fieldClass}
          value={asset}
          onChange={(event) => onAsset(event.target.value as "BRL" | "BTC")}
        >
          <option value="BRL">BRL</option>
          <option value="BTC">BTC</option>
        </select>
        <input
          aria-label={`Valor para ${title}`}
          className={fieldClass}
          value={amount}
          onChange={(event) => onAmount(event.target.value)}
          placeholder="100"
        />
      </div>
      <button className={`${primaryButton} mt-2 w-full`} disabled={disabled}>
        {title}
      </button>
    </form>
  );
}

function LevelList({
  rows,
  tone,
  names,
  selectedId,
  empty,
}: {
  rows: Order[];
  tone: "bid" | "ask";
  names: Map<string, string>;
  selectedId: string;
  empty: string;
}) {
  if (rows.length === 0) return <Empty>{empty}</Empty>;
  return (
    <ul className="flex flex-col gap-1">
      {rows.map((order) => (
        <li key={order.id} className="grid grid-cols-[1fr_auto] items-baseline gap-3 rounded-lg px-2 py-1 font-mono text-xs">
          <span className={tone === "bid" ? "text-bid" : "text-ask"}>
            {formatAmount(order.price)}
            <span className="ml-2 text-white/55">{formatAmount(order.remaining_quantity)}</span>
          </span>
          <span className={order.account_id === selectedId ? "text-gold" : "text-white/40"}>
            {names.get(order.account_id) ?? shortId(order.account_id)}
          </span>
        </li>
      ))}
    </ul>
  );
}

const fieldClass =
  "w-full rounded-xl border border-white/10 bg-black/30 px-3 py-2 text-sm outline-none focus:border-gold/60";
const primaryButton = "rounded-xl bg-white px-3 py-2 text-sm font-medium text-ink disabled:opacity-50";
const ghostButton = "rounded-lg border border-white/15 px-3 py-1 text-xs text-white/80 disabled:opacity-50";
const iconButton = "grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-white text-ink disabled:opacity-50";
