package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/go-flutter-desktop/hover/internal/debugx"
	"github.com/go-flutter-desktop/hover/internal/logx"
	"github.com/go-flutter-desktop/hover/internal/tracex"
	"github.com/spf13/cobra"
)

var (
	verbose int
	colors  bool
	docker  bool
)

func init() {
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "extra logging - more times you declare this the more information is logged")
	rootCmd.PersistentFlags().BoolVar(&colors, "colors", true, "Add colors to log")
	rootCmd.PersistentFlags().BoolVar(&docker, "docker", false, "Run the command in a docker container for hover")
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

func initHover() {
	var (
		colorsOpt = logx.OptionNoop
	)

	if colors {
		colorsOpt = logx.OptionColorize
	}

	switch verbose {
	case 2:
		tracex.Tune(
			colorsOpt,
			logx.OptionStderr,
		)
		tracex.Println("trace logging enabled")
		fallthrough
	case 1:
		debugx.Tune(
			colorsOpt,
			logx.OptionStderr,
		)
		debugx.Println("debug logging enabled")
		fallthrough
	default:
		// informational logging only.
		logx.Tune(colorsOpt)
	}

	tracex.Println("config: colors", colors)
	tracex.Println("config: docker", docker)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println("")
			os.Exit(1)
		}
	}()
}
