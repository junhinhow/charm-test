package main

import (
	"testing"

	"github.com/junhinhow/charm-test"
)

func TestCounter_Increment(t *testing.T) {
	sim := charmtest.New(New())

	// Initial state
	charmtest.RequireViewContains(t, sim, "Count: 0")

	// Press k three times
	sim.SendKeys("k", "k", "k")
	charmtest.RequireViewContains(t, sim, "Count: 3")
}

func TestCounter_Decrement(t *testing.T) {
	sim := charmtest.New(New())

	sim.SendKeys("k", "k", "k") // count = 3
	sim.SendKeys("j", "j")      // count = 1
	charmtest.RequireViewContains(t, sim, "Count: 1")
}

func TestCounter_NoNegative(t *testing.T) {
	sim := charmtest.New(New())

	// Try to go below zero
	sim.SendKeys("j", "j", "j")
	charmtest.RequireViewContains(t, sim, "Count: 0")
}

func TestCounter_Snapshot(t *testing.T) {
	sim := charmtest.New(New())
	sim.SendKeys("k", "k", "k", "k", "k")
	charmtest.RequireSnapshot(t, sim)
}

func TestCounter_Resize(t *testing.T) {
	sim := charmtest.New(New(), charmtest.WithSize(40, 10))
	sim.SendKey("k")
	charmtest.RequireViewContains(t, sim, "Count: 1")
}
