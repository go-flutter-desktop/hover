package cmd

import (
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/config"
	"github.com/go-flutter-desktop/hover/internal/enginecache"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/spf13/cobra"
	"os"
)

var (
	prepareAllTargetOSes bool
	prepareCachePath     string
	prepareEngineVersion string
)

func init() {
	prepareEngineCmd.PersistentFlags().BoolVar(&prepareAllTargetOSes, "all", false, "Prepare the flutter engine for all target operating systems")
	prepareEngineCmd.PersistentFlags().StringVar(&prepareCachePath, "cache-path", enginecache.DefaultCachePath(), "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll")
	prepareEngineCmd.PersistentFlags().StringVar(&prepareEngineVersion, "engine-version", config.BuildEngineDefault, "The flutter engine version to use.")
	prepareEngineCmd.AddCommand(prepareEngineLinuxCmd)
	prepareEngineCmd.AddCommand(prepareEngineDarwinCmd)
	prepareEngineCmd.AddCommand(prepareEngineWindowsCmd)
	rootCmd.AddCommand(prepareEngineCmd)
}

var prepareEngineCmd = &cobra.Command{
	Use:   "prepare-engine",
	Short: "Validates or updates the flutter engine on this machine for a given platform",
	Run: func(cmd *cobra.Command, args []string) {
		initPrepareEngineParameters("darwin", "linux", "windows")
		subcommandPrepare("linux", "windows", "darwin")
	},
}

var prepareEngineLinuxCmd = &cobra.Command{
	Use:   "linux",
	Short: "Validates or updates the flutter engine on this machine for a given platform",
	Run: func(cmd *cobra.Command, args []string) {
		initPrepareEngineParameters("linux")
		subcommandPrepare("linux")
	},
}

var prepareEngineDarwinCmd = &cobra.Command{
	Use:   "darwin",
	Short: "Validates or updates the flutter engine on this machine for a given platform",
	Run: func(cmd *cobra.Command, args []string) {
		initPrepareEngineParameters("darwin")
		subcommandPrepare("darwin")
	},
}

var prepareEngineWindowsCmd = &cobra.Command{
	Use:   "windows",
	Short: "Validates or updates the flutter engine on this machine for a given platform",
	Run: func(cmd *cobra.Command, args []string) {
		initPrepareEngineParameters("windows")
		subcommandPrepare("windows")
	},
}

func initPrepareEngineParameters(targetOSes ...string) {
	if len(targetOSes) == 0 && !prepareAllTargetOSes {
		log.Errorf("missing target OS or --all, cannot continue.")
		os.Exit(1)
	}
}

func subcommandPrepare(targetOSes ...string) {
	for _, targetOS := range targetOSes {
		enginecache.ValidateOrUpdateEngine(targetOS, prepareCachePath, prepareEngineVersion, build.DebugMode)
	}
}
