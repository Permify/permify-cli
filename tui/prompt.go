package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

func StringPrompt(msg string, Placeholder, defaultVal string) (string, error) {
	t := &Tui{}
	prompt := buildPrompt(msg, Placeholder, defaultVal, false)
	t.Inputs = append(t.Inputs, prompt)
	t.Inputs[0].Focus()
	t.Execute()

	if t.Done {
		return t.Inputs[0].Value(), nil
	}

	return "", errors.New("prompt cancelled")
}

func SensitiveStringPrompt(msg string, Placeholder, defaultVal string) (string, error) {
	t := &Tui{}
	prompt := buildPrompt(msg, Placeholder, defaultVal, true)
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
	prompt := buildPrompt(fmt.Sprintf("%s (y/n)", msg), defaultVal, "", false)

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

func buildPrompt(msg string, placeholder string, defaultVal string, sensitive bool) textinput.Model {
	prompt := textinput.New()
	prompt.Prompt = Pink(fmt.Sprintf("%s: ", msg))
	prompt.Placeholder = placeholder
	if defaultVal != "" {
		prompt.SetValue(strings.Trim(defaultVal, "\""))
	}
	if sensitive {
		prompt.EchoMode = textinput.EchoPassword
	}
	return prompt
}
