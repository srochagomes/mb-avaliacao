# Etapas

Modo mentor: o usuário implementa; o agente orienta um artefato por vez e espera o resultado do teste.

Uma etapa por conversa. Ao terminar a etapa, o usuário marca aqui.

## 0 — Contexto local

- [x] Skills, agentes e rule

## 1 — Esqueleto que sobe

- [x] `git init`, remote `https://github.com/srochagomes/mb-avaliacao.git`
- [x] Módulo Go em `back-end/`, `cmd/server`, `GET /health`
- [ ] `docker-compose.yml`: Postgres 16 + API *(adiado — depois do domínio/HTTP)*
- [ ] Migration inicial vazia de schema (tabelas na etapa 3)
- [ ] README com `docker compose up --build` e curl do health

Dev: Go instalado na máquina (`go run`, `go test`). Docker Compose é para a apresentação do recrutador (API + Postgres + front). Domínio e HTTP vêm antes do Compose.

## 2 — Domínio e testes unitários

- [x] Tipos de dinheiro (`Amount`)
- [ ] Tipos de ordem, saldo, book
- [ ] Motor de matching puro (sem I/O), table tests
- [ ] Reserva, liquidação, cancelamento e sobra de preço no domínio

## 3 — Postgres e integração

- [ ] Tabelas, repositórios `pgx`, transação única do place/cancel
- [ ] `pg_advisory_xact_lock` por instrumento
- [ ] Testes de integração contra o Postgres do Compose

## 4 — HTTP

- [ ] Handlers finos, JSON com valores decimais em string
- [ ] Testes de integração HTTP cobrindo o exemplo do enunciado

## 5 — Next.js

- [ ] Telas: saldo, crédito/débito, colocar/cancelar ordem, book
- [ ] Proxy do Next para a API (browser não fala com o Go direto)
- [ ] Serviço no Compose

## 6 — Entrega

- [ ] README do recrutador (subir, usar a UI, curl mínimo)
- [ ] `ASSUMPTIONS.md` curto: prioridade preço-tempo, preço do maker, sem self-trade, escala 8
- [ ] Sobe limpo com `docker compose up --build`
