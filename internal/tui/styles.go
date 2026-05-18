package tui

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

	safeStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399"))
	costlyStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))
	destructiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171"))

	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")).Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3FC")).Bold(true)
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")).Bold(true)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			MarginTop(1)

	confirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F87171")).
			Padding(1, 2).
			MarginTop(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151"))

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7DD3FC")).
			Bold(true)
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
