package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestBuildPromptUsesPasswordEchoModeForSensitiveInputs(t *testing.T) {
	prompt := buildPrompt("enter access token", "", "secret-token", true)

	if got, want := prompt.EchoMode, textinput.EchoPassword; got != want {
		t.Fatalf("EchoMode = %v, want %v", got, want)
	}
	if got, want := prompt.Value(), "secret-token"; got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}

func TestBuildPromptUsesNormalEchoModeForStandardInputs(t *testing.T) {
	prompt := buildPrompt("enter permify url", "", "http://localhost:3478", false)

	if got, want := prompt.EchoMode, textinput.EchoNormal; got != want {
		t.Fatalf("EchoMode = %v, want %v", got, want)
	}
}
