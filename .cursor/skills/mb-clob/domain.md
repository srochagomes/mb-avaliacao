# Domínio fechado

Um book por instrumento (`BTC-BRL`). Vários instrumentos cabem no modelo; o seed é só BTC/BRL.

## Dinheiro

Escala fixa 8. `1` = `100000000` unidades. Aritmética com `math/big.Int`.

Notional em unidades de cotação:

`preco * quantidade / 10^8`

Exemplo do enunciado: compra 1 BTC a 500000 BRL.

- quantidade = 1e8
- preço = 500000e8
- notional = 500000e8

API e UI trafegam string decimal (`"1"`, `"500000"`). Postgres guarda `NUMERIC`. `float` está fora.

Na borda, preço, quantidade e amount de crédito ou débito rejeitam zero, negativo e mais de 8 casas decimais (`ErrInvalidAmount`). A ordem não entra no book. Crédito e débito não movem saldo.

`int64` estoura em `preco * quantidade` nesse exemplo. Por isso `big.Int` (análogo a `BigDecimal`, sem escala flutuante).

## Ordem limitada

Confirmado pela Central de Ajuda (ver [help-center.md](help-center.md)): só executa no preço definido (ou melhor, via matching); sem contraparte fica aberta; **saldo fica bloqueado** até fill ou cancel.

Campos: conta, instrumento, lado (`buy`|`sell`), preço limite, quantidade original, quantidade restante, status (`open`, `filled`, `canceled`), `created_at`.

`open` com restante menor que a original é fill parcial ainda no book.

Fora do exercício: ordem a mercado, stop limit e taxas do produto MB.

## Prioridade e preço

- Compra casa com venda de preço **menor ou igual** ao limite. Melhor venda primeiro (menor preço), depois FIFO (`created_at`, depois id).
- Venda casa com compra de preço **maior ou igual**. Melhor compra primeiro (maior preço), depois FIFO.
- Preço do negócio = preço da ordem que **já estava** no book (maker).
- Sobra de limite volta para disponível. Compra a 510000 que casa a 500000 devolve 10000 da reserva.
- Quantidade que não casar permanece no book, com reserva do restante.
- Conta não casa com a própria ordem. No caminhar do book, pula a ordem da mesma conta e segue na próxima compatível de outra conta. A ordem nova só descansa se, depois desses pulos, não restar liquidez de outra conta.

## Saldo

Por conta e ativo: `available` e `reserved`.

Colocar compra: reserva o notional cheio em BRL (`available -=`, `reserved +=`). Sem disponível suficiente, rejeita e não entra no book.

Colocar venda: reserva a quantidade cheia em BTC.

Negócio: tira da reserva do vendedor o BTC, credita BRL disponível ao vendedor; tira da reserva do comprador o notional ao preço do maker, credita BTC disponível ao comprador. Diferença de preço do comprador volta para `available`.

Cancelar: só ordem `open`. Devolve a reserva do restante. Ordem `filled` ou `canceled` rejeita o cancelamento.

Creditar: soma em `available`. Cria a conta se não existir.

Debitar: só de `available`. Reservado não sai por débito.

## Exemplo que o teste de integração precisa passar

Conta A credita 500000 BRL. Conta B credita 1 BTC.

A compra 1 BTC a 500000. B vende 1 BTC a 500000.

Resultado: A tem 1 BTC e 0 BRL. B tem 0 BTC e 500000 BRL. Book vazio. Um trade de 1 BTC a 500000.
