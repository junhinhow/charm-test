# charm-test

Testing framework for [Bubble Tea](https://github.com/charmbracelet/bubbletea) applications.

Write deterministic, fast tests for your TUI apps — no terminal required.

> **[Leia em Português](README.pt-br.md)**

## The Problem

Testing Bubble Tea apps today means either:
- Manual testing (slow, error-prone, not CI-friendly)
- Writing brittle tests that call Update() manually and inspect model internals
- No snapshot testing, no input simulation, no standard assertions

## The Solution

```go
func TestMyApp(t *testing.T) {
    sim := charmtest.New(NewMyModel())

    sim.Type("hello world")        // simulate typing
    sim.SendKey("enter")           // simulate key press

    charmtest.RequireViewContains(t, sim, "hello world")  // assert on output
    charmtest.RequireSnapshot(t, sim)                      // snapshot testing
}
```

## Installation

```bash
go get github.com/junhinhow/charm-test
```

## Features

### Simulator

The `Simulator` drives a Bubble Tea model without a terminal.

```go
sim := charmtest.New(myModel)
sim := charmtest.New(myModel, charmtest.WithSize(120, 40))
```

Startup follows the real Bubble Tea order: `Init()` is called **before** the
event loop starts, then `tea.ColorProfileMsg`, `tea.WindowSizeMsg` and
`tea.EnvMsg` are delivered to `Update`, in that order, and finally the
command returned by `Init()` is processed.

Options:

| Option | Description |
|---|---|
| `WithSize(w, h)` | Terminal dimensions (must be greater than zero) |
| `WithColorProfile(p)` | Color profile delivered in `tea.ColorProfileMsg` (default: `colorprofile.TrueColor`) |
| `WithEnvironment(env)` | Environment variables delivered in `tea.EnvMsg` (default: `os.Environ()`) |
| `WithMaxMessages(n)` | How many messages are kept for debugging (default: 1000) |
| `WithMaxDepth(n)` | Command recursion limit, protects against command loops (default: 100) |
| `WithMaxEvents(n)` | Total message limit, last resort against models that never settle (default: 100000) |
| `WithMaxAdvanceWait(d)` | How much real time an `Advance` may spend waiting for real timers (default: 10s) |

### Input Simulation

```go
sim.SendKey("enter")              // named keys
sim.SendKey("j")                  // single characters
sim.SendKey("ctrl+c")             // key combinations
sim.SendKeys("k", "k", "k")       // multiple keys
sim.Type("hello")                 // type a string (char by char)
sim.Resize(100, 30)               // simulate terminal resize
```

`SendKey` and `SendKeys` return an `error` when the key name is unknown.
Unknown names never reach the model as arbitrary text — the error is also
recorded in the simulator and reported by every assertion.

**Supported keys:** `enter`, `return`, `esc`/`escape`, `tab`, `shift+tab`,
`backspace`, `delete`, `insert`, `up`, `down`, `left`, `right`, `home`, `end`,
`pgup`/`pageup`, `pgdown`/`pagedown`, `space`, `f1`–`f12`, `capslock`,
`scrolllock`, `numlock`, `printscreen`, `pause`, `menu`, and any single
character (including multibyte ones such as `é` or `漢`).

**Modifiers:** `ctrl`, `alt`, `shift`, `cmd`/`super`, `meta`, `hyper`. They
can be combined: `ctrl+shift+a`, `alt+f1`, `cmd+s`, `shift+enter`.

Uppercase letters behave like a real terminal: `SendKey("H")` produces
`Code='h'`, `Text="H"` and `Mod=ModShift`.

### Virtual Clock

Apps with spinners, clocks or any other timer use `tea.Tick`/`tea.Every`.
Those commands block until their deadline, which would make the simulation
never end. Here they are intercepted and deferred: their message is only
delivered when the virtual clock advances.

```go
sim := charmtest.New(spinnerModel{})   // Init() returns tea.Tick
sim.Advance(3 * time.Second)            // three ticks
sim.AdvanceUntilIdle()                  // deliver everything already ready
sim.Now()                               // current virtual instant
```

`tea.Tick`/`tea.Every` still use the operating system timer underneath, so
delivering them consumes real time (bounded by the advanced window and by
`WithMaxAdvanceWait`; exceeding that budget fails the test with a clear
message instead of hanging). For a fully virtual simulation — no real
waiting at all — use `charmtest.Tick` and `charmtest.Every` in the model
instead:

```go
func (m model) Init() tea.Cmd {
    return charmtest.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

### Assertions

```go
charmtest.RequireView(t, sim, "exact match")
charmtest.RequireViewContains(t, sim, "substring")
charmtest.RequireViewNotContains(t, sim, "should not appear")
charmtest.RequireViewLines(t, sim, 5)  // exactly 5 non-empty lines
```

All assertions strip ANSI escape sequences before comparison and normalize
`\r\n` line endings, so the same test works on every operating system.

### Snapshot Testing

```go
charmtest.RequireSnapshot(t, sim)                       // auto-named by test
charmtest.RequireSnapshotNamed(t, sim, "after-login")   // custom name
```

Golden files live in `testdata/`. A snapshot covers the **full view** —
content, alternate screen, cursor, colors, progress bar and window title —
not just the rendered text.

A missing golden file **fails the test**: a new snapshot is never accepted
silently in CI.

**Create or update snapshots:** `CHARM_TEST_UPDATE=1 go test ./...`

### Debugging

```go
charmtest.DumpView(t, sim)        // print current view to test log
charmtest.DumpMessages(t, sim)    // print all processed messages
charmtest.ViewString(sim)         // get ANSI-stripped view as string
```

### Model State Access

```go
model, ok := charmtest.ModelAs[MyModel](sim)   // works with value receivers
if !ok {
    t.Fatal("unexpected model")
}
if model.Count != 5 {
    t.Errorf("got %d, want 5", model.Count)
}
```

`ModelAs` is the safe way to reach the model: models with a **value**
receiver are stored by value, so `sim.Model().(*MyModel)` panics.
`sim.Model()` is still available for models with a pointer receiver.

### View Inspection

```go
view := sim.FullView()        // the complete tea.View
snap := sim.ViewSnapshot()    // serializable snapshot of the view
if snap.AltScreen { ... }
if snap.Cursor != nil && snap.Cursor.X == 0 { ... }
```

## Full Example

See [`examples/counter/`](examples/counter/) for a complete working example with tests.

```go
func TestCounter_Increment(t *testing.T) {
    sim := charmtest.New(counter.New())

    if err := sim.SendKeys("k", "k", "k"); err != nil {
        t.Fatal(err)
    }
    charmtest.RequireViewContains(t, sim, "Count: 3")
}

func TestCounter_Snapshot(t *testing.T) {
    sim := charmtest.New(counter.New())
    if err := sim.SendKeys("k", "k", "k", "k", "k"); err != nil {
        t.Fatal(err)
    }
    charmtest.RequireSnapshot(t, sim)
}
```

## Design Principles

- **Deterministic:** the same input always produces the same output. Commands
  run synchronously; the only exception is Bubble Tea's own real timers
  (`tea.Tick`/`tea.Every`), which are deferred and delivered when the virtual
  clock advances
- **Safe:** unknown key names, invalid sizes and command loops fail loudly
  instead of hanging or crashing the test binary
- **Fast:** tests run in microseconds, except when the model uses real timers
- **Simple:** one import, one constructor, intuitive API
- **Idiomatic:** works with `testing.T`, follows Go testing conventions
- **ANSI-aware:** all comparisons strip escape sequences automatically

## License

MIT
