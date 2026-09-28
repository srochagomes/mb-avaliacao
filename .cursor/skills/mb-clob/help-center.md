# Fontes da Central de Ajuda MB

Extraído em 2026-09-25. Só o que importa para o CLOB simplificado. UI do app/site e taxas reais do MB **não** entram na entrega.

Fontes:
- [Ordem limitada](https://central.ajuda.mercadobitcoin.com.br/wiki/spaces/TF/pages/490111092)
- [Guia para iniciantes](https://central.ajuda.mercadobitcoin.com.br/wiki/spaces/TF/pages/560824400)

## Ordem limitada (artigo 490111092)

- Compra ou vende **somente** quando o mercado atinge o preço que o usuário definiu.
- Fica ativa no book até executar (total/parcial) ou cancelar.
- Exige contraparte compatível; sem oferta correspondente, permanece aberta.
- **Saldo usado na ordem fica indisponível** até execução ou cancelamento → no domínio: `available` → `reserved`.

Passo a passo de tela do MB e faixa de taxa 0,3–0,7%: fora do escopo do exercício.

## Guia iniciantes — trechos úteis (artigo 560824400)

**Negociação:** o MB não é a contraparte; clientes criam ordens de compra/venda. Tipos no produto real: mercado, limitada, stop limit. **Neste exercício só limitada.**

**Limitada (resumo do guia):**
- Usuário define o preço.
- Executa quando o mercado atinge esse valor (e há contraparte).
- Em limitada e stop limit, **saldo permanece bloqueado** até executar ou cancelar.

**Livro de ordens:**
- Ordem nova entra no book aguardando oferta correspondente.
- Combinação exige uma compra e uma venda com valores compatíveis.

**Preço / mercado / stop:** documentação de produto; market e stop **não** implementamos.

## De-para para o domínio fechado

| Ajuda MB | Nossa decisão |
|---|---|
| Saldo bloqueado na limitada | `reserved` na place; libera no fill/cancel |
| Precisa contraparte compatível | match por preço-tempo; resto fica `open` |
| Livro de compra e venda | bids + asks por instrumento |
| Taxas 0,3–0,7% | **não** modelar (CLOB simplificado) |
| Ordem a mercado / stop | **não** implementar |
