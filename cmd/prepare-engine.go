package cmd

import (
	"os"
	"runtime"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/config"
	"github.com/go-flutter-desktop/hover/internal/enginecache"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/spf13/cobra"
)

var (
	prepareCachePath     string
	prepareEngineVersion string
	prepareReleaseMode   bool
	prepareDebugMode     bool
	prepareProfileMode   bool
	prepareBuildModes    []build.Mode
)

func init() {
	prepareEngineCmd.PersistentFlags().StringVar(&prepareCachePath, "cache-path", enginecache.DefaultCachePath(), "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll")
	prepareEngineCmd.PersistentFlags().StringVar(&prepareEngineVersion, "engine-version", config.BuildEngineDefault, "The flutter engine version to use.")
	prepareEngineCmd.PersistentFlags().BoolVar(&prepareDebugMode, "debug", false, "Prepare the flutter engine for debug mode")
	prepareEngineCmd.PersistentFlags().BoolVar(&prepareReleaseMode, "release", false, "Prepare the flutter engine for release mode.")
	prepareEngineCmd.PersistentFlags().BoolVar(&prepareProfileMode, "profile", false, "Prepare the flutter engine for profile mode.")
	prepareEngineCmd.AddCommand(prepareEngineLinuxCmd)
	prepareEngineCmd.AddCommand(prepareEngineDarwinCmd)
	prepareEngineCmd.AddCommand(prepareEngineWindowsCmd)
	rootCmd.AddCommand(prepareEngineCmd)
}

var prepareEngineCmd = &cobra.Command{
	Use:   "prepare-engine",
	Short: "Validates or updates the flutter engine on this machine for a given platform",
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

func initPrepareEngineParameters(targetOS string) {
	validatePrepareEngineParameters(targetOS)
	if prepareDebugMode {
		prepareBuildModes = append(prepareBuildModes, build.DebugMode)
	}
	if prepareReleaseMode {
		prepareBuildModes = append(prepareBuildModes, build.ReleaseMode)
	}
	if prepareProfileMode {
		prepareBuildModes = append(prepareBuildModes, build.ProfileMode)
	}
}

func validatePrepareEngineParameters(targetOS string) {
	numberOfPrepareModeFlagsSet := 0
	for _, flag := range []bool{prepareProfileMode, prepareProfileMode, prepareDebugMode} {
		if flag {
			numberOfPrepareModeFlagsSet++
		}
	}
	if numberOfPrepareModeFlagsSet > 1 {
		log.Errorf("Only one of --debug, --release or --profile can be set at one time")
		os.Exit(1)
	}
	if numberOfPrepareModeFlagsSet == 0 {
		prepareDebugMode = true
	}
	if targetOS == "darwin" && runtime.GOOS != targetOS && (prepareReleaseMode || prepareProfileMode) {
		log.Errorf("It is not possible to prepare the flutter engine in release mode for darwin using docker")
		os.Exit(1)
	}
}

func subcommandPrepare(targetOS string) {
	for _, mode := range prepareBuildModes {
		enginecache.ValidateOrUpdateEngine(targetOS, prepareCachePath, prepareEngineVersion, mode)
	}
}
