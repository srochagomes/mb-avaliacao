# API

Prefixo `/v1`. Valores monetários são string decimal. Ids são UUID gerados no Postgres.

| Método | Caminho | Corpo / resposta |
|---|---|---|
| GET | `/health` | `{"status":"ok"}` |
| POST | `/v1/accounts/{id}/credits` | `{"asset":"BRL","amount":"500000"}` |
| POST | `/v1/accounts/{id}/debits` | `{"asset":"BTC","amount":"1"}` |
| GET | `/v1/accounts/{id}/balances` | `[{"asset":"BRL","available":"500000","reserved":"0"}]` |
| POST | `/v1/orders` | `{"account_id","instrument":"BTC-BRL","side":"buy","price":"500000","quantity":"1"}` |
| DELETE | `/v1/orders/{id}` | ordem cancelada |
| GET | `/v1/books/BTC-BRL` | cada ordem aberta, sem agregar por preço |

`POST /v1/orders` devolve a ordem e os trades gerados nesse place (para o recrutador ver o match sem ler o banco).

`GET /v1/books/BTC-BRL` lista cada ordem aberta. Bids em preço decrescente e, no mesmo preço, FIFO (`created_at`, depois id). Asks em preço crescente e FIFO. Cada item traz `id`, `price` e `remaining`.

Instrumento na URL usa hífen (`BTC-BRL`).

O Next faz proxy de `/backend/*` para o serviço Go. O browser não chama o host da API. CORS não é necessário.
