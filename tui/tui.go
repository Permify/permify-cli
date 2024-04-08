package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)


type TextInput struct {
	Msg         string
	Placeholder string
	IntialValue string
	FinalValue  string
}

type Tui struct {
	Inputs  []textinput.Model
	focused int
	Err     error
	Done    bool
}

func (t *Tui) Execute() {
	p := tea.NewProgram(t)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
	}

}

func (m *Tui) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.Inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.Inputs)-1 {
				m.Done = true
				return m, tea.Quit
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.Inputs {
			m.Inputs[i].Blur()
		}
		m.Inputs[m.focused].Focus()

	// We handle errors just like any other message
	case error:
		m.Err = msg
		return m, nil
	}

	for i := range m.Inputs {
		m.Inputs[i], cmds[i] = m.Inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m *Tui) View() string {
	output := strings.Builder{}
	// Write input to screen
	for i, k := range m.Inputs {
		//output.WriteString(k.View())
		if i < m.focused {
			output.WriteString(k.View())
		}
		if i == m.focused {
			output.WriteString(k.View())
		}
	}
	// add any final message
	output.WriteString("\n")
	return output.String()
}

// nextInput focuses the next input field
func (m *Tui) nextInput() {
	m.focused = (m.focused + 1) % len(m.Inputs)
}

// prevInput focuses the previous input field
func (m *Tui) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.Inputs) - 1
	}
}
