// Package logger defines the logging functions
package logger

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	clog "github.com/charmbracelet/log"
)

var Log = clog.New()

func Update(debugEnabled bool) {
	clog.ErrorLevelStyle.SetString("ERROR")
	clog.FatalLevelStyle.SetString("FATAL")

	if debugEnabled {
		Log.SetReportCaller(true)
		Log.SetLevel(clog.DebugLevel)
	}
}

func init() {
	clog.ErrorLevelStyle = lipgloss.NewStyle().
		SetString(strings.ToUpper("ERROR")).
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "203",
			Dark:  "204",
		})

	clog.FatalLevelStyle = lipgloss.NewStyle().
		SetString(strings.ToUpper("FATAL")).
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "133",
			Dark:  "134",
		})
	clog.DebugLevelStyle = lipgloss.NewStyle().
		SetString(strings.ToUpper("DEBUG")).
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "63",
			Dark:  "63",
		})
	Log.SetReportTimestamp(true)
	Log.SetTimeFormat(time.Kitchen)
}
