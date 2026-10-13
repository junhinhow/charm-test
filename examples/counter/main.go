// Exemplo: um contador simples para demonstrar o charm-test.
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// Model é um contador simples que incrementa com "k"/up e decrementa com
// "j"/down.
type Model struct {
	Count int
	quit  bool
}

// New cria um contador zerado.
func New() Model {
	return Model{}
}

// Init implementa tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implementa tea.Model.
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

// View implementa tea.Model.
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
