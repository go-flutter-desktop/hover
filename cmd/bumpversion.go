package cmd

import (
	"fmt"
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
	Use:   "bumpversion",
	Short: "upgrade the 'go-flutter' golang library in this project",
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterProject()
		// Hardcode target to the current OS (no cross-compile for this command)
		targetOS := runtime.GOOS

		err := upgrade(targetOS)
		if err != nil {
			os.Exit(1)
		}
	},
}

func upgrade(targetOS string) (err error) {
	var engineCachePath string
	if buildCachePath != "" {
		engineCachePath = enginecache.ValidateOrUpdateEngineAtPath(targetOS, buildCachePath, "")
	} else {
		engineCachePath = enginecache.ValidateOrUpdateEngine(targetOS, "")
	}
	return upgradeGoFlutter(targetOS, engineCachePath)
}

func upgradeGoFlutter(targetOS string, engineCachePath string) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Errorf("Failed to get working dir: %v", err)
		return
	}

	var cgoLdflags string
	switch targetOS {
	case "darwin":
		cgoLdflags = fmt.Sprintf("-F%s -Wl,-rpath,@executable_path", engineCachePath)
	case "linux":
		cgoLdflags = fmt.Sprintf("-L%s", engineCachePath)
	case "windows":
		cgoLdflags = fmt.Sprintf("-L%s", engineCachePath)
	default:
		log.Errorf("Target platform %s is not supported, cgo_ldflags not implemented.", targetOS)
		return
	}

	if buildBranch == "" {
		buildBranch = "@latest"
	}

	cmdGoGetU := exec.Command(build.GoBin, "get", "-u", "github.com/go-flutter-desktop/go-flutter"+buildBranch)
	cmdGoGetU.Dir = filepath.Join(wd, build.BuildPath)
	cmdGoGetU.Env = append(os.Environ(),
		"GOPROXY=direct", // github.com/golang/go/issues/32955 (allows '/' in branch name)
		"GO111MODULE=on",
		"CGO_LDFLAGS="+cgoLdflags,
	)
	cmdGoGetU.Stderr = os.Stderr
	cmdGoGetU.Stdout = os.Stdout

	err = cmdGoGetU.Run()
	// When cross-compiling the command fails, but that is not an error
	if err != nil && !buildDocker {
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

	log.Printf("'go-flutter' is on version: %s", currentTag)

	return nil

}
