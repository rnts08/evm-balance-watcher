package tui

import "github.com/charmbracelet/lipgloss"

// --- Styles ---
var (
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle  = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)
	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	boxStyle  = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1)
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Bold(true).
				Padding(0, 1)
)
