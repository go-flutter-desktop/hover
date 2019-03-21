package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hover",
	Short: "Hover connects Flutter and go-flutter-desktop.",
	Long:  `Hover helps developers to release Flutter applications on desktop.`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

// Execute executes the rootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("hover command failed: %v\n", err)
		os.Exit(1)
	}
}
