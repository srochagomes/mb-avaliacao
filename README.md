# Livro BTC/BRL · BTC/BRL Order Book

<p align="center">
  <strong>PT-BR</strong> · Simulação de livro de ofertas (CLOB) em memória<br/>
  <strong>EN</strong> · In-memory central limit order book simulation
</p>

<p align="center">
  API em <strong>Go</strong> · Tela em <strong>Next.js</strong> · Sobe com <strong>Docker Compose</strong>
</p>

---

# Português

## O que é isto

Uma simulação simplificada de livro de ofertas **BTC/BRL**: você cria carteiras, credita saldo, coloca ordens limitadas, cancela e vê o matching acontecer ao vivo.

- Preço do negócio = preço de quem **já estava** no livro (maker)
- Prioridade **preço-tempo**
- Sem self-trade (a mesma carteira não casa consigo)
- Estado **em memória** — reiniciar a API zera carteiras, saldos e o livro

| Serviço | URL |
|---|---|
| Tela | http://localhost:3001 |
| API | http://localhost:8081 |
| Saúde | `curl -s localhost:8081/health` |

---

## Pré-requisitos

### Opção A — Docker (recomendado para avaliar a entrega)

Instale o [Docker Engine](https://docs.docker.com/engine/install/) e o [Docker Compose](https://docs.docker.com/compose/install/) (plugin `docker compose`).

Confira:

```bash
docker --version
docker compose version
```

### Opção B — Go local (só para testes unitários)

Para rodar `go test` sem Docker:

1. Instale o [Go 1.22+](https://go.dev/dl/)
2. Confira: `go version`

A tela e a API em produção de demo sobem pelo Compose. O Go local é suficiente para validar o domínio e a API em memória.

---

## Subir a aplicação

Na raiz do repositório:

```bash
chmod +x scripts/up.sh scripts/down.sh   # só na primeira vez
./scripts/up.sh
```

Aguarde o build terminar. Quando aparecer `Ready`, abra:

**http://localhost:3001**

| Ação | Comando |
|---|---|
| Subir (primeiro plano) | `./scripts/up.sh` |
| Parar com `Ctrl+C` | derruba containers e a rede |
| Subir em segundo plano | `./scripts/up.sh -d` |
| Parar (segundo plano) | `./scripts/down.sh` |

---

## A tela em três abas

### 1. Carteiras e saldo

| Elemento | O que faz |
|---|---|
| Campo + botão **+** | Cria uma carteira (ex.: `Conta A`) |
| Lista de carteiras | Seleciona qual conta você está operando |
| Cartões **BRL** / **BTC** | Mostram **disponível** e **reservado** |
| **Creditar** | Soma no disponível (ativo + valor) |
| **Debitar** | Só tira do disponível (reservado não sai) |
| Roteiro do enunciado | Atalho do exemplo oficial |

### 2. Ordens

| Elemento | O que faz |
|---|---|
| **Compra** / **Venda** | Lado da ordem limitada |
| Preço limite (BRL) | Preço máximo de compra ou mínimo de venda |
| Quantidade (BTC) | Quanto você quer negociar |
| **Colocar ordem** | Reserva saldo, tenta casar, resto fica no livro |
| Lista de ordens | Histórico da carteira selecionada |
| **Cancelar** | Só em ordem aberta; devolve a reserva |

### 3. Livro e negócios

| Elemento | O que faz |
|---|---|
| **Última negociação** (topo) | Preço do último trade |
| Asks / Bids | Vendas e compras abertas no livro |
| Histórico | Lista de negócios casados |

---

## Passo a passo — exemplo do enunciado

### Passo 1 · Criar duas carteiras

1. Abra a aba **Carteiras e saldo**
2. Digite `Conta A` e clique em **+**
3. Digite `Conta B` e clique em **+**
4. Clique em cada carteira na lista para selecioná-la

### Passo 2 · Creditar saldo

1. Selecione **Conta A**
2. Em **Creditar**, escolha ativo `BRL`, valor `500000`, clique em **Creditar**
3. Selecione **Conta B**
4. Em **Creditar**, escolha ativo `BTC`, valor `1`, clique em **Creditar**

### Passo 3 · Conta A compra

1. Mantenha **Conta A** selecionada
2. Vá à aba **Ordens**
3. Escolha **Compra**, preço `500000`, quantidade `1`
4. Clique em **Colocar ordem**

Resultado esperado: ordem **aberta**, `500000` BRL **reservados**, oferta aparece no livro.

### Passo 4 · Conta B vende

1. Volte a **Carteiras e saldo**, selecione **Conta B**
2. Aba **Ordens** → **Venda**, preço `500000`, quantidade `1`
3. Clique em **Colocar ordem**

Resultado esperado:

| Carteira | BTC | BRL |
|---|---|---|
| Conta A | 1 disponível | 0 |
| Conta B | 0 | 500000 disponível |

O livro fica vazio. O campo **Última negociação** mostra `500000 BRL`.

### Outras operações úteis

- **Cancelar**: ordem aberta na aba Ordens → **Cancelar** (reserva volta ao disponível)
- **Sobra de preço**: venda a `500000` no livro; compra a `510000` — o negócio sai a `500000` e `10000` BRL voltam ao disponível do comprador
- **Débito**: só afeta o disponível; se o valor estiver reservado, a operação falha com feedback na tela

---

## Testes (Go local)

```bash
cd back-end
go test ./...
```

Cobre domínio (matching, reserva, liquidação) e o fluxo HTTP do enunciado em memória.

---

## Premissas

Ver [`ASSUMPTIONS.md`](ASSUMPTIONS.md).

---

# English

## What this is

A simplified **BTC/BRL** central limit order book: create wallets, credit balances, place limit orders, cancel, and watch matching live.

- Trade price = resting order price (**maker**)
- **Price-time** priority
- No self-trade
- **In-memory** state — restarting the API clears everything

| Service | URL |
|---|---|
| UI | http://localhost:3001 |
| API | http://localhost:8081 |
| Health | `curl -s localhost:8081/health` |

---

## Prerequisites

### Option A — Docker (recommended for reviewers)

Install [Docker Engine](https://docs.docker.com/engine/install/) and [Docker Compose](https://docs.docker.com/compose/install/).

```bash
docker --version
docker compose version
```

### Option B — Local Go (unit tests only)

1. Install [Go 1.22+](https://go.dev/dl/)
2. Check: `go version`

The demo UI/API run via Compose. Local Go is enough to validate the domain and in-memory HTTP tests.

---

## Run the app

From the repository root:

```bash
chmod +x scripts/up.sh scripts/down.sh   # first time only
./scripts/up.sh
```

When the build finishes and you see `Ready`, open:

**http://localhost:3001**

| Action | Command |
|---|---|
| Start (foreground) | `./scripts/up.sh` |
| Stop with `Ctrl+C` | tears down containers and network |
| Start in background | `./scripts/up.sh -d` |
| Stop (background) | `./scripts/down.sh` |

---

## The UI in three tabs

### 1. Wallets & balance

| Control | Purpose |
|---|---|
| Field + **+** | Create a wallet (e.g. `Account A`) |
| Wallet list | Select the active account |
| **BRL** / **BTC** cards | **Available** and **reserved** balances |
| **Credit** | Adds to available |
| **Debit** | Removes from available only (not reserved) |
| Spec walkthrough | Shortcut for the official example |

### 2. Orders

| Control | Purpose |
|---|---|
| **Buy** / **Sell** | Limit order side |
| Limit price (BRL) | Max buy / min sell price |
| Quantity (BTC) | Size to trade |
| **Place order** | Reserves funds, matches, rests leftover |
| Order list | Orders for the selected wallet |
| **Cancel** | Open orders only; releases reservation |

### 3. Book & trades

| Control | Purpose |
|---|---|
| **Last trade** (header) | Last traded price |
| Asks / Bids | Open sell and buy levels |
| History | Matched trades |

---

## Walkthrough — problem statement example

### Step 1 · Create two wallets

1. Open **Wallets & balance**
2. Type `Account A`, click **+**
3. Type `Account B`, click **+**
4. Click a wallet in the list to select it

### Step 2 · Fund balances

1. Select **Account A**
2. **Credit** → asset `BRL`, amount `500000`
3. Select **Account B**
4. **Credit** → asset `BTC`, amount `1`

### Step 3 · Account A buys

1. Keep **Account A** selected
2. Open **Orders**
3. **Buy**, price `500000`, quantity `1` → **Place order**

Expected: order **open**, `500000` BRL **reserved**, bid on the book.

### Step 4 · Account B sells

1. Select **Account B** under **Wallets & balance**
2. **Orders** → **Sell**, price `500000`, quantity `1` → **Place order**

Expected:

| Wallet | BTC | BRL |
|---|---|---|
| Account A | 1 available | 0 |
| Account B | 0 | 500000 available |

Book is empty. **Last trade** shows `500000 BRL`.

### Other useful flows

- **Cancel**: open order → **Cancel** (reservation returns to available)
- **Price surplus**: resting sell at `500000`, buy at `510000` — trade at `500000`, buyer gets `10000` BRL back
- **Debit**: available only; reserved funds reject the debit with an on-screen error

---

## Tests (local Go)

```bash
cd back-end
go test ./...
```

Covers domain (matching, reserve, settlement) and the in-memory HTTP scenario from the problem statement.

---

## Assumptions

See [`ASSUMPTIONS.md`](ASSUMPTIONS.md).
