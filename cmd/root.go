package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/go-flutter-desktop/hover/internal/logx"
	"github.com/spf13/cobra"
)

var colors bool
var docker bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&colors, "colors", true, "Add colors to log")
	rootCmd.PersistentFlags().BoolVar(&docker, "docker", false, "Run the command in a docker container for hover")
}

func initHover() {
	if colors {
		logx.Tune(logx.OptionColorize)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println("")
			os.Exit(1)
		}
	}()
}

var rootCmd = &cobra.Command{
	Use:   "hover",
	Short: "Hover connects Flutter and go-flutter-desktop.",
	Long:  "Hover helps developers to release Flutter applications on desktop.",
}

// Execute executes the rootCmd
func Execute() {
	cobra.OnInitialize(initHover)
	if err := rootCmd.Execute(); err != nil {
		logx.Errorf("Command failed: %v", err)
		os.Exit(1)
	}
}
