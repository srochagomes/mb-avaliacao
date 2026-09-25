# Premissas

- Prioridade preço-tempo: melhor preço primeiro, depois `created_at`, depois id.
- O preço do negócio é o da ordem que já estava no livro (maker).
- Conta não casa com a própria ordem.
- Dinheiro em escala 8, com `big.Int`. A API trafega decimal em string.
- Só ordem limitada em `BTC-BRL`. Sem ordem a mercado, stop ou taxa.
- O estado fica na memória de um processo. Não há Postgres nesta entrega.
