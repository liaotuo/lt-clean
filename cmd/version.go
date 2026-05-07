package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set via -ldflags at build time.
var Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the lt-clean version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("lt-clean", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
