package cmd

import (
	"fmt"
	"github.com/go-flutter-desktop/hover/internal/log"
	"os"
	"os/signal"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var au aurora.Aurora
var colors bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&colors, "colors", true, "Add colors to log")
	cobra.OnInitialize(initColors)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println("")
			os.Exit(1)
		}
	}()

}

func initColors() {
	au = aurora.NewAurora(colors)
	log.Au = au
}

var rootCmd = &cobra.Command{
	Use:   "hover",
	Short: "Hover connects Flutter and go-flutter-desktop.",
	Long:  "Hover helps developers to release Flutter applications on desktop.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

// Execute executes the rootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Command failed: %v", err)
		os.Exit(1)
	}
}
