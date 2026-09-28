---
name: go-from-java
description: De-para Java para Go, memória, convenções e testes. Use ao escrever ou explicar código Go neste repositório para quem vem de Java.
---

# Go para quem vem de Java

Ensinar em uma frase no momento em que o conceito aparece. Não transformar a resposta em aula.

## Mapa

| Java | Go neste projeto |
|---|---|
| classe + campos | `struct`; método com receiver `func (o Order) Remaining()` |
| interface explícita | interface implícita: se tem os métodos, implementa |
| interface no `ports` | interface declarada no pacote `app`, implementada em `postgres` |
| exceção | `error` no retorno; `fmt.Errorf("...: %w", err)` |
| `Optional` | ponteiro, ou `valor, ok` |
| `null` | `nil` |
| `ArrayList` | slice `[]Order` |
| `HashMap` | `map[K]V` |
| `synchronized` | `sync.Mutex` no processo; advisory lock no Postgres entre processos |
| `try-with-resources` | `defer rows.Close()` |
| `public` / pacote | `Nome` exporta; `nome` fica no pacote |
| `BigDecimal` | `math/big.Int` com escala 8 |
| JUnit parametrizado | `testing` + table test (`t.Run`) |
| injeção Spring | `main` monta structs e passa dependências |
| anotações | struct tag `` `json:"amount"` `` |
| Stream | `for` |

## Memória

- Slice é ponteiro + len + cap para um array. Atribuir o slice não copia os elementos. Copiar com `slices.Clone` quando o book não pode compartilhar backing array.
- `map` é referência. Passar o map deixa o callee alterar o original.
- Struct pequena vai por valor. Ponteiro quando o método altera o valor ou o struct é grande e sai do `app` para o banco.
- O compilador decide stack ou heap (escape analysis). Não há `new` obrigatório. `new` e `&T{}` só alocam.
- Goroutine sem dono vaza. Este serviço não dispara goroutine por request; o `net/http` já faz isso.
- Variável e import sem uso não compilam.
- String é imutável. Concatenar em loop aloca; aqui os JSON são pequenos, `encoding/json` basta.
- `defer` roda na saída da função, em ordem inversa. `defer` dentro de `for` segura o recurso até o fim da função; fechar `rows` no bloco, não acumular defer no loop de várias queries.

## Convenções que o código segue

- `gofmt` é a formatação. Sem discussão de estilo.
- Erros em minúsculas, sem ponto final: `insufficient balance`.
- `context.Context` é o primeiro parâmetro de função que faz I/O.
- Aceitar interface, devolver struct.
- Nome de pacote curto, sem stutter: `order.Order` seria ruim; o tipo mora em `domain` como `domain.Order`.
- Siglas: `ID`, `URL`.
- Teste ao lado: `money_test.go`. Integração com build tag `//go:build integration`.
- Handler devolve cedo no erro. Sem `else` depois de `return`.

## Testes

Table test:

```go
tests := []struct {
    name string
    // ...
}{}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { /* ... */ })
}
```

`go test ./...` é o unitário. `go test -tags=integration ./...` sobe contra o Postgres do Compose.
