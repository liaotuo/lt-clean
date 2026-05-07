package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lt-clean",
	Short: "Mac dev-environment cleaner — TUI + CLI",
	Long: "lt-clean scans common dev caches, IDE artifacts, mobile build dirs,\n" +
		"system logs, and project leftovers, then lets you reclaim disk space.\n\n" +
		"Run with no arguments to launch the interactive TUI.",
	RunE: func(cmd *cobra.Command, args []string) error {
		items := catalog.Build()
		p := tea.NewProgram(tui.New(items), tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

// Execute is the entrypoint called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
