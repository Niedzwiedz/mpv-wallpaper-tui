package ui

import "github.com/charmbracelet/lipgloss"

var (
	orange = lipgloss.Color("#ffa07a")
	muted  = lipgloss.Color("240")
	black  = lipgloss.Color("0")

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(muted)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(orange)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Background(orange).
			Foreground(black).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(muted)

	dimStyle = lipgloss.NewStyle().
			Foreground(muted)
)
