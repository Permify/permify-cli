package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// var ErrorStyle = lipgloss.NewStyle().
// 	Foreground(lipgloss.Color("#FFFDF5")).
// 	Background(lipgloss.Color("#fc0349")).
// 	Padding(0).
// 	MarginRight(1)

// func Println(stsyle lipgloss.Style, msg string) {
// 	fmt.Println(stsyle.Render(msg))
// }

// Blue returns a colored string of #8df9d9
func Blue(str string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#8DF9D9")).Render(str)
}

// Pink returns a colored string of #ff06b7
func Pink(str string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF06B7")).Render(str)
}

// Warning returns a colored string of #ff6a69
func Warning(str string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6A69")).Render(str)
}

// Critical returns a colored string of #ff3933
func Critical(str string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF3933")).Render(str)
}
