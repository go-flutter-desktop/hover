package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/internal/version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Hover version information",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("No arguments allowed")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		version := version.HoverVersion()
		fmt.Printf("Hover %s\n", version)
	},
}
