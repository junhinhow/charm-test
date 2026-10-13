# charm-test

Framework de testes para aplicações [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Escreva testes determinísticos e rápidos para seus apps TUI — sem precisar de terminal.

> **[Read in English](README.md)**

## O Problema

Testar apps Bubble Tea hoje significa:
- Testes manuais (lento, propenso a erros, não funciona em CI)
- Testes frágeis que chamam Update() manualmente e inspecionam internals do model
- Sem snapshot testing, sem simulação de input, sem assertions padrão

## A Solução

```go
func TestMeuApp(t *testing.T) {
    sim := charmtest.New(NewMeuModel())

    sim.Type("hello world")        // simula digitação
    sim.SendKey("enter")           // simula tecla

    charmtest.RequireViewContains(t, sim, "hello world")  // assert no output
    charmtest.RequireSnapshot(t, sim)                      // snapshot testing
}
```

## Instalação

```bash
go get github.com/junhinhow/charm-test
```

## Funcionalidades

### Simulador

O `Simulator` roda um model Bubble Tea sem terminal.

```go
sim := charmtest.New(meuModel)
sim := charmtest.New(meuModel, charmtest.WithSize(120, 40))
```

A inicialização segue a mesma ordem do Bubble Tea real: `Init()` é chamado
**antes** do event loop começar; depois, `tea.ColorProfileMsg`,
`tea.WindowSizeMsg` e `tea.EnvMsg` são entregues ao `Update`, nessa ordem; e,
por fim, o comando devolvido por `Init()` é processado.

Opções:

| Opção | Descrição |
|---|---|
| `WithSize(w, h)` | Dimensões do terminal (precisam ser maiores que zero) |
| `WithColorProfile(p)` | Perfil de cores entregue em `tea.ColorProfileMsg` (padrão: `colorprofile.TrueColor`) |
| `WithEnvironment(env)` | Variáveis de ambiente entregues em `tea.EnvMsg` (padrão: `os.Environ()`) |
| `WithMaxMessages(n)` | Quantas mensagens são guardadas para depuração (padrão: 1000) |
| `WithMaxDepth(n)` | Limite de recursão de comandos, protege contra loops de comandos (padrão: 100) |

### Simulação de Input

```go
sim.SendKey("enter")              // teclas nomeadas
sim.SendKey("j")                  // caracteres individuais
sim.SendKey("ctrl+c")             // combinações de tecla
sim.SendKeys("k", "k", "k")       // multiplas teclas
sim.Type("hello")                 // digita uma string (char por char)
sim.Resize(100, 30)               // simula resize do terminal
```

`SendKey` e `SendKeys` retornam um `error` quando o nome da tecla é
desconhecido. Nomes desconhecidos nunca chegam ao modelo como texto
arbitrário — o erro também fica registrado no simulador e é reportado por
qualquer assertion.

**Teclas suportadas:** `enter`, `return`, `esc`/`escape`, `tab`, `shift+tab`,
`backspace`, `delete`, `insert`, `up`, `down`, `left`, `right`, `home`, `end`,
`pgup`/`pageup`, `pgdown`/`pagedown`, `space`, `f1`–`f12`, `capslock`,
`scrolllock`, `numlock`, `printscreen`, `pause`, `menu`, e qualquer caractere
individual (inclusive multibyte, como `é` ou `漢`).

**Modificadores:** `ctrl`, `alt`, `shift`, `cmd`/`super`, `meta`, `hyper`.
Podem ser combinados: `ctrl+shift+a`, `alt+f1`, `cmd+s`, `shift+enter`.

Letras maiúsculas se comportam como em um terminal real: `SendKey("H")`
produz `Code='h'`, `Text="H"` e `Mod=ModShift`.

### Relógio Virtual

Apps com spinner, relógio ou qualquer outro temporizador usam
`tea.Tick`/`tea.Every`. Esses comandos bloqueiam até o prazo, o que faria a
simulação nunca terminar. Aqui eles são interceptados e adiados: a mensagem só
é entregue quando o relógio virtual avança.

```go
sim := charmtest.New(spinnerModel{})   // Init() devolve tea.Tick
sim.Advance(3 * time.Second)            // três ticks
sim.AdvanceUntilIdle()                  // entrega o que já está pronto
sim.Now()                               // instante atual do relógio virtual
```

`tea.Tick`/`tea.Every` continuam usando o timer do sistema operacional por
baixo dos panos, portanto entregá-los consome tempo real (limitado pela
janela avançada). Para uma simulação 100% virtual — sem nenhuma espera real
— use `charmtest.Tick` e `charmtest.Every` no modelo:

```go
func (m model) Init() tea.Cmd {
    return charmtest.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

### Assertions

```go
charmtest.RequireView(t, sim, "match exato")
charmtest.RequireViewContains(t, sim, "substring")
charmtest.RequireViewNotContains(t, sim, "nao deve aparecer")
charmtest.RequireViewLines(t, sim, 5)  // exatamente 5 linhas nao-vazias
```

Todas as assertions removem sequências ANSI antes de comparar e normalizam
as quebras de linha `\r\n`, então o mesmo teste funciona em todos os sistemas
operacionais.

### Snapshot Testing

```go
charmtest.RequireSnapshot(t, sim)                       // nome automatico pelo teste
charmtest.RequireSnapshotNamed(t, sim, "apos-login")    // nome customizado
```

Arquivos golden ficam em `testdata/`. O snapshot cobre a view **completa** —
conteúdo, tela alternativa, cursor, cores, barra de progresso e título da
janela — e não apenas o texto renderizado.

Um golden inexistente **falha o teste**: um snapshot novo nunca é aceito
silenciosamente no CI.

**Criar ou atualizar snapshots:** `CHARM_TEST_UPDATE=1 go test ./...`

### Debugging

```go
charmtest.DumpView(t, sim)        // imprime view atual no log de teste
charmtest.DumpMessages(t, sim)    // imprime todas mensagens processadas
charmtest.ViewString(sim)         // view sem ANSI como string
```

### Acesso ao State do Model

```go
model, ok := charmtest.ModelAs[MeuModel](sim)   // funciona com receptor por valor
if !ok {
    t.Fatal("modelo inesperado")
}
if model.Count != 5 {
    t.Errorf("obtive %d, esperava 5", model.Count)
}
```

`ModelAs` é a forma segura de acessar o model: modelos com receptor **por
valor** são guardados por valor, então `sim.Model().(*MeuModel)` entra em
pânico. `sim.Model()` continua disponível para modelos com receptor por
ponteiro.

### Inspeção da View

```go
view := sim.FullView()        // a tea.View completa
snap := sim.ViewSnapshot()    // retrato serializável da view
if snap.AltScreen { ... }
if snap.Cursor != nil && snap.Cursor.X == 0 { ... }
```

## Exemplo Completo

Veja [`examples/counter/`](examples/counter/) para um exemplo funcional com testes.

```go
func TestContador_Incrementar(t *testing.T) {
    sim := charmtest.New(counter.New())

    if err := sim.SendKeys("k", "k", "k"); err != nil {
        t.Fatal(err)
    }
    charmtest.RequireViewContains(t, sim, "Count: 3")
}

func TestContador_Snapshot(t *testing.T) {
    sim := charmtest.New(counter.New())
    if err := sim.SendKeys("k", "k", "k", "k", "k"); err != nil {
        t.Fatal(err)
    }
    charmtest.RequireSnapshot(t, sim)
}
```

## Princípios de Design

- **Determinístico:** a mesma entrada sempre produz a mesma saída. Os comandos
  rodam de forma síncrona; a única exceção são os temporizadores reais do
  próprio Bubble Tea (`tea.Tick`/`tea.Every`), que são adiados e entregues
  quando o relógio virtual avança
- **Seguro:** nomes de tecla desconhecidos, tamanhos inválidos e loops de
  comandos falham de forma explícita, em vez de travar ou derrubar o binário
  de testes
- **Rápido:** os testes rodam em microssegundos, exceto quando o model usa
  temporizadores reais
- **Simples:** um import, um construtor, API intuitiva
- **Idiomático:** funciona com `testing.T`, segue convenções de teste do Go
- **ANSI-aware:** todas comparações removem escape sequences automaticamente

## Licença

MIT
