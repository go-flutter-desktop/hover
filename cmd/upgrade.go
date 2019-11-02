package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/enginecache"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/versioncheck"
)

func init() {
	upgradeCmd.Flags().StringVarP(&buildCachePath, "cache-path", "", "", "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll (defaults to the standard user cache directory)")
	upgradeCmd.Flags().MarkHidden("branch")
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade the `go-flutter` core library",
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterProject()
		// Can only run on host OS
		buildTarget := build.Target{
			Platform:        runtime.GOOS,
			PackagingFormat: "",
		}
		err := upgrade(buildTarget)
		if err != nil {
			os.Exit(1)
		}
	},
}

func upgrade(buildTarget build.Target) (err error) {
	var engineCachePath string
	if buildCachePath != "" {
		engineCachePath = enginecache.ValidateOrUpdateEngineAtPath(buildTarget, buildCachePath)
	} else {
		engineCachePath = enginecache.ValidateOrUpdateEngine(buildTarget)
	}
	return upgradeGoFlutter(buildTarget, engineCachePath)
}

func upgradeGoFlutter(buildTarget build.Target, engineCachePath string) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Errorf("Failed to get working dir: %v", err)
		return
	}

	if buildBranch == "" {
		buildBranch = "@latest"
	}

	cmdGoGetU := exec.Command(build.GoBin, "get", "-u", "github.com/go-flutter-desktop/go-flutter"+buildBranch)
	cmdGoGetU.Dir = filepath.Join(wd, build.BuildPath)
	cmdGoGetU.Env = append(os.Environ(),
		"GOPROXY=direct", // github.com/golang/go/issues/32955 (allows `/` in branch name)
		"GO111MODULE=on",
		"CGO_LDFLAGS="+build.CGoLdFlags(buildTarget, engineCachePath),
	)
	cmdGoGetU.Stderr = os.Stderr
	cmdGoGetU.Stdout = os.Stdout

	err = cmdGoGetU.Run()
	if err != nil {
		log.Errorf("Updating go-flutter to %s version failed: %v", buildBranch, err)
		return
	}

	cmdGoModDownload := exec.Command(build.GoBin, "mod", "download")
	cmdGoModDownload.Dir = filepath.Join(wd, build.BuildPath)
	cmdGoModDownload.Env = append(os.Environ(),
		"GO111MODULE=on",
	)
	cmdGoModDownload.Stderr = os.Stderr
	cmdGoModDownload.Stdout = os.Stdout

	err = cmdGoModDownload.Run()
	if err != nil {
		log.Errorf("Go mod download failed: %v", err)
		return
	}

	currentTag, err := versioncheck.CurrentGoFlutterTag(filepath.Join(wd, build.BuildPath))
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

	log.Printf("`go-flutter` is on version: %s", currentTag)

	return nil

}
