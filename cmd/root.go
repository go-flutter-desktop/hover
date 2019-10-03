package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/spf13/cobra"
)

var colors bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&colors, "colors", true, "Add colors to log")
}

func initHover() {
	if colors {
		log.Colorize()
	}
	initBinaries()
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
		log.Errorf("Command failed: %v", err)
		os.Exit(1)
	}
}
