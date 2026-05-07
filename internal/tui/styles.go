package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7DD3FC")).
			Padding(0, 1)

	groupStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A78BFA")).
			MarginTop(1)

	safeStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399"))
	costlyStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))
	destructiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171"))

	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")).Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3FC")).Bold(true)
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")).Bold(true)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1)

	confirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F87171")).
			Padding(1, 2).
			MarginTop(1)
)

func levelStyle(level string) lipgloss.Style {
	switch level {
	case "Safe":
		return safeStyle
	case "Costly":
		return costlyStyle
	case "Destructive":
		return destructiveStyle
	}
	return dimStyle
}
