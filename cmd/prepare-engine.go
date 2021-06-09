package cmd

import (
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/config"
	"github.com/go-flutter-desktop/hover/internal/enginecache"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"runtime"
)

var (
	prepareCachePath     string
	prepareEngineVersion string
	prepareReleaseMode   bool
	prepareDebugMode     bool
	prepareBuildModes    []build.Mode
)

func init() {
	prepareEngineCmd.PersistentFlags().StringVar(&prepareCachePath, "cache-path", enginecache.DefaultCachePath(), "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll")
	prepareEngineCmd.PersistentFlags().StringVar(&prepareEngineVersion, "engine-version", config.BuildEngineDefault, "The flutter engine version to use.")
	prepareEngineCmd.PersistentFlags().BoolVar(&prepareDebugMode, "debug-mode", true, "Prepare the flutter engine for debug mode")
	prepareEngineCmd.PersistentFlags().BoolVar(&prepareReleaseMode, "release-mode", false, "Prepare the flutter engine for release mode")
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
	if prepareDebugMode {
		prepareBuildModes = append(prepareBuildModes, build.DebugMode)
	}
	if prepareReleaseMode {
		prepareBuildModes = append(prepareBuildModes, build.ReleaseMode)
	}
	validatePrepareEngineParameters(targetOS)
}

func validatePrepareEngineParameters(targetOS string) {
	if targetOS == "darwin" && runtime.GOOS != targetOS && prepareReleaseMode {
		if path, err := exec.LookPath("darling"); err != nil || len(path) == 0 {
			log.Errorf("To prepare the release flutter engine for darwin on linux, install darling from your package manager or https://www.darlinghq.org/")
			os.Exit(1)
		}
	}
}

func subcommandPrepare(targetOS string) {
	for _, mode := range prepareBuildModes {
		enginecache.ValidateOrUpdateEngine(targetOS, prepareCachePath, prepareEngineVersion, mode)
	}
}
