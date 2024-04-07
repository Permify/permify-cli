package tui

import (
	"fmt"
	"strings"

	"github.com/Permify/permify-cli/core/logger"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	cursor  int
	choice  string
	prompt  string
	choices []string // Add a field for the choices
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			logger.Log.Fatal("prompt was cancelled by user")

		case "enter":
			// Send the choice on the channel and exit.
			m.choice = m.choices[m.cursor]
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}
	// s.WriteString(m.prompt)
	s.WriteString(Pink(fmt.Sprintf("%s: ", m.prompt)))
	s.WriteString("\n\n")

	for i := 0; i < len(m.choices); i++ {
		if m.cursor == i {
			s.WriteString(Blue("[•]"))
		} else {
			s.WriteString(Blue("[ ] "))
		}
		s.WriteString(m.choices[i])
		s.WriteString("\n")
	}
	return s.String()
}

func Choice(prompt string, choices []string) (string, error) {
	p := tea.NewProgram(model{
		prompt:  prompt,
		choices: choices, // Pass the choices to the model
	})

	// Run returns the model as a tea.Model.
	m, err := p.Run()
	if err != nil {
		return "", err
	}
	// Assert the final tea.Model to our local model and print the choice.
	if m, ok := m.(model); ok && m.choice != "" {
		return m.choice, nil
	}
	return "", nil
}
