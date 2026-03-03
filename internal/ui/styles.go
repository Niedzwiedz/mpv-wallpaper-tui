package ui

import "github.com/charmbracelet/lipgloss"

var (
	accent = lipgloss.Color("2")   // ANSI green
	muted  = lipgloss.Color("240")
	black  = lipgloss.Color("0")

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(muted)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// selectedStyle: cursor is here AND this section is focused.
	selectedStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Background(accent).
			Foreground(black).
			Bold(true)

	// activeStyle: item is selected but this section is not focused
	// (e.g. the chosen monitor while navigating wallpapers).
	activeStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(accent).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(muted)

	dimStyle = lipgloss.NewStyle().
			Foreground(muted)
)

// InitStyles applies optional colour overrides from config and rebuilds all
// style vars. Must be called before creating the Model.
func InitStyles(overrideAccent, overrideMuted string) {
	if overrideAccent != "" {
		accent = lipgloss.Color(overrideAccent)
	}
	if overrideMuted != "" {
		muted = lipgloss.Color(overrideMuted)
	}
	// Rebuild derived styles so they pick up the updated colours.
	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(muted)
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(accent)
	selectedStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Background(accent).
		Foreground(black).
		Bold(true)
	activeStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(accent).
		Bold(true)
	helpStyle = lipgloss.NewStyle().
		Foreground(muted)
	dimStyle = lipgloss.NewStyle().
		Foreground(muted)
}
