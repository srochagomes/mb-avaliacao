# Arquitetura

Camadas curtas, no estilo que o mercado Go usa. Interfaces ficam em quem consome (o serviço), não num pacote `ports`.

```
back-end/
  cmd/server/main.go          # composição (o main do Spring Boot)
  internal/domain/            # entidades, dinheiro, matching puro
  internal/app/               # casos de uso; declara as interfaces de repositório
  internal/httpapi/           # handlers e JSON
  internal/postgres/          # pgx
  migrations/
front-end/                    # Next.js App Router
docker-compose.yml
README.md
```

## SOLID neste desenho

- Um caso de uso por arquivo (`PlaceOrder`, `CancelOrder`, `Credit`, `Debit`, `GetBalances`, `GetBook`).
- Domínio não importa `net/http` nem `pgx`.
- Handler não calcula match nem saldo.
- Repositório Postgres é detalhe; o app depende da interface.
- Matching é função pura testável sem banco.

## Transação do place

Um `BEGIN` faz tudo: trava o instrumento (`pg_advisory_xact_lock`), trava saldos (`SELECT … FOR UPDATE`), reserva, lê o lado oposto, casa, grava trades e ordens, liquida, commit. Cancelamento é outra transação no mesmo estilo.

Dois processos não casam o mesmo book ao mesmo tempo por causa do advisory lock (o `synchronized` distribuído).

## Erros de negócio

Sentinela no domínio (`ErrInsufficientBalance`, `ErrOrderNotOpen`, `ErrInvalidAmount`). HTTP mapeia para 422. Falha de infra é 500. Validação de JSON é 400.

## O que não entra

Fila, cache, auth, websocket, framework HTTP, ORM, lib de decimal, float.
