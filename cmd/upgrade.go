package cmd

import (
	"fmt"
	"github.com/go-flutter-desktop/hover/internal/build"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/go-flutter-desktop/hover/internal/enginecache"
	"github.com/go-flutter-desktop/hover/internal/versioncheck"

	"github.com/spf13/cobra"
)

func init() {
	upgradeCmd.Flags().StringVarP(&buildCachePath, "cache-path", "", "", "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll (defaults to the standard user cache directory)")
	upgradeCmd.Flags().MarkHidden("branch")
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade the 'go-flutter' core library",
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterProject()
		// Hardcode target to the current OS (no cross-compile support yet)
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
		engineCachePath = enginecache.ValidateOrUpdateEngineAtPath(targetOS, buildCachePath)
	} else {
		engineCachePath = enginecache.ValidateOrUpdateEngine(targetOS)
	}
	return upgradeGoFlutter(targetOS, engineCachePath)
}

func upgradeGoFlutter(targetOS string, engineCachePath string) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("hover: Failed to get working dir: %v\n", err)
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
		fmt.Printf("hover: Target platform %s is not supported, cgo_ldflags not implemented.\n", targetOS)
		return
	}

	cmdGoGetU := exec.Command(build.GoBin, "get", "-u", "github.com/go-flutter-desktop/go-flutter"+buildBranch)
	cmdGoGetU.Dir = filepath.Join(wd, build.BuildPath)
	cmdGoGetU.Env = append(os.Environ(),
		"GO111MODULE=on",
		"CGO_LDFLAGS="+cgoLdflags,
	)
	cmdGoGetU.Stderr = os.Stderr
	cmdGoGetU.Stdout = os.Stdout

	err = cmdGoGetU.Run()
	if err != nil {
		versionName := buildBranch
		if versionName == "" {
			versionName = "latest"
		}
		fmt.Printf("hover: Updating go-flutter to %s version failed: %v\n", versionName, err)
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
		fmt.Printf("hover: Go mod download failed: %v\n", err)
		return
	}

	currentTag, err := versioncheck.CurrentGoFlutterTag(filepath.Join(wd, build.BuildPath))
	if err != nil {
		fmt.Printf("hover: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("hover: 'go-flutter' is on version: %s\n", currentTag)

	return nil

}
