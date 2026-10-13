package main

import (
	"testing"

	"github.com/junhinhow/charm-test"
)

func TestCounter_Increment(t *testing.T) {
	sim := charmtest.New(New())

	// Estado inicial
	charmtest.RequireViewContains(t, sim, "Count: 0")

	// Pressiona k três vezes
	if err := sim.SendKeys("k", "k", "k"); err != nil {
		t.Fatalf("falha ao enviar as teclas: %v", err)
	}
	charmtest.RequireViewContains(t, sim, "Count: 3")
}

func TestCounter_Decrement(t *testing.T) {
	sim := charmtest.New(New())

	if err := sim.SendKeys("k", "k", "k"); err != nil { // count = 3
		t.Fatalf("falha ao enviar as teclas: %v", err)
	}
	if err := sim.SendKeys("j", "j"); err != nil { // count = 1
		t.Fatalf("falha ao enviar as teclas: %v", err)
	}
	charmtest.RequireViewContains(t, sim, "Count: 1")
}

func TestCounter_NoNegative(t *testing.T) {
	sim := charmtest.New(New())

	// Tenta chegar abaixo de zero
	if err := sim.SendKeys("j", "j", "j"); err != nil {
		t.Fatalf("falha ao enviar as teclas: %v", err)
	}
	charmtest.RequireViewContains(t, sim, "Count: 0")
}

func TestCounter_Snapshot(t *testing.T) {
	sim := charmtest.New(New())
	if err := sim.SendKeys("k", "k", "k", "k", "k"); err != nil {
		t.Fatalf("falha ao enviar as teclas: %v", err)
	}
	charmtest.RequireSnapshot(t, sim)
}

func TestCounter_Resize(t *testing.T) {
	sim := charmtest.New(New(), charmtest.WithSize(40, 10))
	if err := sim.SendKey("k"); err != nil {
		t.Fatalf("falha ao enviar a tecla: %v", err)
	}
	charmtest.RequireViewContains(t, sim, "Count: 1")
}

// TestCounter_ModelAs mostra como acessar o estado de um modelo com receptor
// por valor (o tipo guardado pelo simulador é o valor, não o ponteiro).
func TestCounter_ModelAs(t *testing.T) {
	sim := charmtest.New(New())
	if err := sim.SendKey("k"); err != nil {
		t.Fatalf("falha ao enviar a tecla: %v", err)
	}

	model, ok := charmtest.ModelAs[Model](sim)
	if !ok {
		t.Fatal("não foi possível converter o modelo")
	}
	if model.Count != 1 {
		t.Errorf("esperava Count=1, obtive %d", model.Count)
	}
}
