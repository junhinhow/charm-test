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
The `Simulator` drives a Bubble Tea model synchronously — no goroutines, no terminal, fully deterministic.

```go
sim := charmtest.New(myModel)
sim := charmtest.New(myModel, charmtest.WithSize(120, 40))
```

### Input Simulation
```go
sim.SendKey("enter")              // named keys
sim.SendKey("j")                  // single characters
sim.SendKey("ctrl+c")             // key combinations
sim.SendKeys("k", "k", "k")      // multiple keys
sim.Type("hello")                 // type a string (char by char)
sim.Resize(100, 30)               // simulate terminal resize
```

**Supported keys:** `enter`, `esc`, `tab`, `shift+tab`, `backspace`, `delete`, `up`, `down`, `left`, `right`, `home`, `end`, `pgup`, `pgdown`, `space`, `ctrl+c`, `ctrl+a`, `ctrl+d`, `ctrl+e`, `ctrl+k`, `ctrl+u`, `ctrl+n`, `ctrl+p`, and any single character.

### Assertions
```go
charmtest.RequireView(t, sim, "exact match")
charmtest.RequireViewContains(t, sim, "substring")
charmtest.RequireViewNotContains(t, sim, "should not appear")
charmtest.RequireViewLines(t, sim, 5)  // exactly 5 non-empty lines
```

All assertions strip ANSI escape sequences before comparison, so styled output is matched by content only.

### Snapshot Testing
```go
charmtest.RequireSnapshot(t, sim)                       // auto-named by test
charmtest.RequireSnapshotNamed(t, sim, "after-login")   // custom name
```

Golden files are stored in `testdata/`. On first run, the snapshot is created. On subsequent runs, the view is compared against the stored snapshot.

**Update snapshots:** `CHARM_TEST_UPDATE=1 go test ./...`

### Debugging
```go
charmtest.DumpView(t, sim)        // print current view to test log
charmtest.DumpMessages(t, sim)    // print all processed messages
charmtest.ViewString(sim)         // get ANSI-stripped view as string
```

### Model State Access
```go
model := sim.Model().(*MyModel)   // type assert to your model
assert.Equal(t, 5, model.Count)
```

## Full Example

See [`examples/counter/`](examples/counter/) for a complete working example with tests.

```go
func TestCounter_Increment(t *testing.T) {
    sim := charmtest.New(counter.New())

    sim.SendKeys("k", "k", "k")
    charmtest.RequireViewContains(t, sim, "Count: 3")
}

func TestCounter_Snapshot(t *testing.T) {
    sim := charmtest.New(counter.New())
    sim.SendKeys("k", "k", "k", "k", "k")
    charmtest.RequireSnapshot(t, sim)
}
```

## Design Principles

- **Deterministic:** No goroutines, no timers — same input always produces same output
- **Fast:** Tests run in microseconds, not seconds
- **Simple:** One import, one constructor, intuitive API
- **Idiomatic:** Works with `testing.T`, follows Go testing conventions
- **ANSI-aware:** All comparisons strip escape sequences automatically

## License

MIT
