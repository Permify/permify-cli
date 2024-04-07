package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

func StringPrompt(msg string, Placeholder, defaultVal string) (string, error) {
	t := &Tui{}
	prompt := textinput.New()
	prompt.Prompt = Pink(fmt.Sprintf("%s: ", msg))
	prompt.Placeholder = Placeholder
	if defaultVal != "" {
		prompt.SetValue(strings.Trim(defaultVal, "\""))
	}
	t.Inputs = append(t.Inputs, prompt)
	t.Inputs[0].Focus()
	t.Execute()

	if t.Done {
		return t.Inputs[0].Value(), nil
	}

	return "", errors.New("prompt cancelled")
}

func BoolPrompt(msg string, defaultVal string) (bool, error) {
	t := &Tui{}
	prompt := textinput.New()
	prompt.Prompt = Pink(fmt.Sprintf("%s (y/n): ", msg))
	prompt.Placeholder = defaultVal

	t.Inputs = append(t.Inputs, prompt)
	t.Execute()

	if t.Done {
		if t.Inputs[0].Value() == "y" {
			return true, nil
		} else if t.Inputs[0].Value() == "n" {
			return false, nil
		} else {
			return false, errors.New("invalid input")
		}
	}
	return false, errors.New("prompt cancelled")
}
