package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"

	"github.com/go-flutter-desktop/hover/internal/logx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
		version := hoverVersion()
		fmt.Printf("Hover %s\n", version)
	},
}

var (
	hoverVersionValue string
	hoverVersionOnce  sync.Once
)

func hoverVersion() string {
	hoverVersionOnce.Do(func() {
		buildInfo, ok := debug.ReadBuildInfo()
		if !ok {
			logx.Errorf("Cannot obtain version information from hover build. To resolve this, please go-get hover using Go 1.13 or newer.")
			os.Exit(1)
		}
		hoverVersionValue = buildInfo.Main.Version
	})
	return hoverVersionValue
}
