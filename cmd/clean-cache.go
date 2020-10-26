package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/internal/enginecache"
	"github.com/go-flutter-desktop/hover/internal/log"
)

var cachePath string

func init() {
	cleanCacheCmd.Flags().StringVar(&cachePath, "cache-path", enginecache.DefaultCachePath(), "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll")
	rootCmd.AddCommand(cleanCacheCmd)
}

var cleanCacheCmd = &cobra.Command{
	Use:   "clean-cache",
	Short: "Clean cached engine files",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("does not take arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := os.RemoveAll(enginecache.BaseEngineCachePath(cachePath))
		if err != nil {
			log.Errorf("Failed to delete engine cache directory: %v", err)
			os.Exit(1)
		}
	},
}
