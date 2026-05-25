package findtui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7DD3FC")).
		Background(lipgloss.Color("#1E3A5F")).
		Padding(0, 2).
		MarginBottom(1)

	groupStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#A78BFA")).
		Background(lipgloss.Color("#2D1B4E")).
		Padding(0, 1).
		MarginTop(1).
		MarginBottom(0)

	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")).Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3FC")).Bold(true)

	barStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#374151"))
	sizeBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3FC"))
)
