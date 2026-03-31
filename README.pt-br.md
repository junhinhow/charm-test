# charm-test

Framework de testes para aplicacoes [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Escreva testes deterministicos e rapidos para seus apps TUI — sem precisar de terminal.

> **[Read in English](README.md)**

## O Problema

Testar apps Bubble Tea hoje significa:
- Testes manuais (lento, propenso a erros, nao funciona em CI)
- Testes frageis que chamam Update() manualmente e inspecionam internals do model
- Sem snapshot testing, sem simulacao de input, sem assertions padrao

## A Solucao

```go
func TestMeuApp(t *testing.T) {
    sim := charmtest.New(NewMeuModel())

    sim.Type("hello world")        // simula digitacao
    sim.SendKey("enter")           // simula tecla

    charmtest.RequireViewContains(t, sim, "hello world")  // assert no output
    charmtest.RequireSnapshot(t, sim)                      // snapshot testing
}
```

## Instalacao

```bash
go get github.com/junhinhow/charm-test
```

## Funcionalidades

### Simulador
O `Simulator` roda um model Bubble Tea de forma sincrona — sem goroutines, sem terminal, totalmente deterministico.

```go
sim := charmtest.New(meuModel)
sim := charmtest.New(meuModel, charmtest.WithSize(120, 40))
```

### Simulacao de Input
```go
sim.SendKey("enter")              // teclas nomeadas
sim.SendKey("j")                  // caracteres individuais
sim.SendKey("ctrl+c")             // combinacoes de tecla
sim.SendKeys("k", "k", "k")      // multiplas teclas
sim.Type("hello")                 // digita uma string (char por char)
sim.Resize(100, 30)               // simula resize do terminal
```

**Teclas suportadas:** `enter`, `esc`, `tab`, `shift+tab`, `backspace`, `delete`, `up`, `down`, `left`, `right`, `home`, `end`, `pgup`, `pgdown`, `space`, `ctrl+c`, `ctrl+a`, `ctrl+d`, `ctrl+e`, `ctrl+k`, `ctrl+u`, `ctrl+n`, `ctrl+p`, e qualquer caractere individual.

### Assertions
```go
charmtest.RequireView(t, sim, "match exato")
charmtest.RequireViewContains(t, sim, "substring")
charmtest.RequireViewNotContains(t, sim, "nao deve aparecer")
charmtest.RequireViewLines(t, sim, 5)  // exatamente 5 linhas nao-vazias
```

Todas as assertions removem sequencias ANSI antes de comparar, entao output estilizado e comparado apenas pelo conteudo.

### Snapshot Testing
```go
charmtest.RequireSnapshot(t, sim)                       // nome automatico pelo teste
charmtest.RequireSnapshotNamed(t, sim, "apos-login")    // nome customizado
```

Arquivos golden sao armazenados em `testdata/`. Na primeira execucao, o snapshot e criado. Nas seguintes, a view e comparada com o snapshot armazenado.

**Atualizar snapshots:** `CHARM_TEST_UPDATE=1 go test ./...`

### Debugging
```go
charmtest.DumpView(t, sim)        // imprime view atual no log de teste
charmtest.DumpMessages(t, sim)    // imprime todas mensagens processadas
charmtest.ViewString(sim)         // view sem ANSI como string
```

### Acesso ao State do Model
```go
model := sim.Model().(*MeuModel)  // type assert pro seu model
assert.Equal(t, 5, model.Count)
```

## Exemplo Completo

Veja [`examples/counter/`](examples/counter/) para um exemplo funcional com testes.

```go
func TestContador_Incrementar(t *testing.T) {
    sim := charmtest.New(counter.New())

    sim.SendKeys("k", "k", "k")
    charmtest.RequireViewContains(t, sim, "Count: 3")
}

func TestContador_Snapshot(t *testing.T) {
    sim := charmtest.New(counter.New())
    sim.SendKeys("k", "k", "k", "k", "k")
    charmtest.RequireSnapshot(t, sim)
}
```

## Principios de Design

- **Deterministico:** Sem goroutines, sem timers — mesmo input sempre produz mesmo output
- **Rapido:** Testes rodam em microsegundos, nao segundos
- **Simples:** Um import, um construtor, API intuitiva
- **Idiomatico:** Funciona com `testing.T`, segue convencoes de teste do Go
- **ANSI-aware:** Todas comparacoes removem escape sequences automaticamente

## Licenca

MIT
