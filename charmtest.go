// Package charmtest provides a testing framework for Bubble Tea applications.
//
// It simulates the Bubble Tea runtime without requiring a real terminal,
// allowing you to send messages, simulate key presses, and assert on the
// rendered output and model state.
//
// Basic usage:
//
//	func TestMyModel(t *testing.T) {
//	    m := NewMyModel()
//	    sim := charmtest.New(m)
//	    sim.SendKey("j")           // simulate pressing "j"
//	    sim.SendKey("enter")       // simulate pressing enter
//	    charmtest.RequireView(t, sim, "expected output")
//	}
package charmtest

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// DefaultWidth is the default terminal width for simulations.
const DefaultWidth = 80

// DefaultHeight is the default terminal height for simulations.
const DefaultHeight = 24

// Simulator drives a Bubble Tea model through its lifecycle without a real
// terminal. It processes messages synchronously, making tests deterministic.
type Simulator struct {
	model  tea.Model
	width  int
	height int

	// messages collects all messages processed for debugging.
	messages []tea.Msg
}

// Option configures a Simulator.
type Option func(*Simulator)

// WithSize sets the terminal dimensions for the simulation.
func WithSize(width, height int) Option {
	return func(s *Simulator) {
		s.width = width
		s.height = height
	}
}

// New creates a Simulator that wraps the given Bubble Tea model.
// The model's Init() is called and resulting commands are processed.
func New(model tea.Model, opts ...Option) *Simulator {
	s := &Simulator{
		model:  model,
		width:  DefaultWidth,
		height: DefaultHeight,
	}
	for _, opt := range opts {
		opt(s)
	}

	// Send initial WindowSizeMsg
	s.Send(tea.WindowSizeMsg{Width: s.width, Height: s.height})

	// Process Init commands
	cmd := s.model.Init()
	s.processCmd(cmd)

	return s
}

// Send sends a message to the model and processes the resulting command.
func (s *Simulator) Send(msg tea.Msg) {
	s.messages = append(s.messages, msg)
	var cmd tea.Cmd
	s.model, cmd = s.model.Update(msg)
	s.processCmd(cmd)
}

// SendKey simulates a key press. Accepts key names like "enter", "esc",
// "tab", "up", "down", "ctrl+c", or single characters like "a", "j".
func (s *Simulator) SendKey(key string) {
	s.Send(keyMsg(key))
}

// Type simulates typing a string, sending one key press per character.
func (s *Simulator) Type(text string) {
	for _, ch := range text {
		s.SendKey(string(ch))
	}
}

// SendKeys simulates a sequence of key presses.
func (s *Simulator) SendKeys(keys ...string) {
	for _, k := range keys {
		s.SendKey(k)
	}
}

// View returns the current rendered output of the model.
func (s *Simulator) View() string {
	return s.model.View().Content
}

// Model returns the current model state. Use type assertion to access
// your specific model type:
//
//	m := sim.Model().(*MyModel)
func (s *Simulator) Model() tea.Model {
	return s.model
}

// Messages returns all messages that were sent to the model.
func (s *Simulator) Messages() []tea.Msg {
	return s.messages
}

// Resize simulates a terminal resize event.
func (s *Simulator) Resize(width, height int) {
	s.width = width
	s.height = height
	s.Send(tea.WindowSizeMsg{Width: width, Height: height})
}

// ViewContains returns true if the current view contains the given substring.
func (s *Simulator) ViewContains(substr string) bool {
	return strings.Contains(s.View(), substr)
}

// processCmd executes a command and sends the resulting message back.
// Batch commands are processed recursively. Async commands (that return
// tea.Msg from a function) are executed synchronously for determinism.
func (s *Simulator) processCmd(cmd tea.Cmd) {
	if cmd == nil {
		return
	}

	msg := cmd()
	if msg == nil {
		return
	}

	// Handle batch messages
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			s.processCmd(c)
		}
		return
	}

	// Skip quit messages in tests
	if _, ok := msg.(tea.QuitMsg); ok {
		return
	}

	s.Send(msg)
}
