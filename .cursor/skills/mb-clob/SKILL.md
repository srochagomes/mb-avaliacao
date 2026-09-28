---
name: mb-clob
description: Regras fechadas do exercício Mercado Bitcoin (CLOB, matching, saldos, API, camadas, etapas). Use ao implementar, revisar ou explicar o back-end Go, o livro de ofertas, ordens, saldos, Postgres, testes ou o README desta avaliação.
---

# CLOB — contexto fechado

Não releia `spec/`. Este skill é a fonte.

## Ler sob demanda

- Etapa atual e o que falta: [stages.md](stages.md)
- Invariantes de ordem, match e saldo: [domain.md](domain.md)
- Pastas, SOLID e transação: [architecture.md](architecture.md)
- Contrato HTTP: [api.md](api.md)
- Links do PDF (Central de Ajuda): [help-center.md](help-center.md)
- De-para Java e memória: [../go-from-java/SKILL.md](../go-from-java/SKILL.md)

## O que a avaliação pede

Obrigatório: colocar ordem limitada, cancelar ordem, matching, saldos coerentes com o match.

Bônus que esta entrega inclui: creditar, debitar, consultar saldo, consultar o book.

Também pedimos (além do enunciado): testes unitários, testes de integração, Next.js, Postgres, Docker Compose, README de um comando.

Preferir stdlib. Justificar no README a única lib de produção (`pgx`) e as decisões de domínio.
