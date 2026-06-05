// Example: a simple counter model for demonstrating charm-test.
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// Model is a simple counter that increments with "k"/up and decrements with "j"/down.
type Model struct {
	Count int
	quit  bool
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "k", "up":
			m.Count++
		case "j", "down":
			if m.Count > 0 {
				m.Count--
			}
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	if m.quit {
		return tea.View{}
	}
	return tea.View{Content: fmt.Sprintf("Count: %d\n\nPress k/j to change, q to quit.", m.Count)}
}

func main() {
	p := tea.NewProgram(New())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
