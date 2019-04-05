package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/otiai10/copy"
	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/internal/enginecache"
)

var dotSlash = string([]byte{'.', filepath.Separator})

var buildTargetMainDart string

func init() {
	buildCmd.Flags().StringVarP(&buildTargetMainDart, "target", "t", "lib/main_desktop.dart", "The main entry-point file of the application.")
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a desktop release",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := assertInFlutterProject()
		assertHoverInitialized()

		// Hardcode target to the current OS (no cross-compile support yet)
		targetOS := runtime.GOOS

		build(projectName, targetOS, nil)
	},
}

func build(projectName string, targetOS string, vmArguments []string) {

	// TODO: release build
	// "--disable-dart-asserts", // release mode flag
	// "--disable-observatory",

	outputDirectoryPath, err := filepath.Abs(filepath.Join("desktop", "build", "outputs", targetOS))
	if err != nil {
		fmt.Printf("Failed to resolve absolute path for output directory: %v\n", err)
		os.Exit(1)
	}

	var outputBinaryName = projectName
	switch targetOS {
	case "darwin":
		// no special filename
	case "linux":
		// no special filename
	case "windows":
		outputBinaryName += ".exe"
	default:
		fmt.Printf("Target platform %s is not supported.\n", targetOS)
		os.Exit(1)
	}
	outputBinaryPath := filepath.Join(outputDirectoryPath, outputBinaryName)

	engineCachePath := enginecache.ValidateOrUpdateEngine(targetOS)

	err = os.RemoveAll(outputDirectoryPath)
	if err != nil {
		fmt.Printf("failed to clean output directory %s: %v\n", outputDirectoryPath, err)
		os.Exit(1)
	}

	err = os.MkdirAll(outputDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("failed to create output directory %s: %v\n", outputDirectoryPath, err)
		os.Exit(1)
	}

	cmdFlutterBuild := exec.Command(flutterBin, "build", "bundle",
		"--asset-dir", filepath.Join(outputDirectoryPath, "flutter_assets"),
		"--target", buildTargetMainDart,
	)
	cmdFlutterBuild.Stderr = os.Stderr
	cmdFlutterBuild.Stdout = os.Stdout
	err = cmdFlutterBuild.Run()
	if err != nil {
		fmt.Printf("Flutter build failed: %v\n", err)
		os.Exit(1)
	}

	var engineFile string
	switch targetOS {
	case "darwin":
		engineFile = "FlutterEmbedder.framework"
	case "linux":
		engineFile = "libflutter_engine.so"
	case "windows":
		engineFile = "flutter_engine.dll"
	}

	err = copy.Copy(
		filepath.Join(engineCachePath, engineFile),
		filepath.Join(outputDirectoryPath, engineFile),
	)
	if err != nil {
		fmt.Printf("Failed to copy %s: %v\n", engineFile, err)
		os.Exit(1)
	}

	err = copy.Copy(
		filepath.Join(engineCachePath, "artifacts", "icudtl.dat"),
		filepath.Join(outputDirectoryPath, "icudtl.dat"),
	)
	if err != nil {
		fmt.Printf("Failed to copy icudtl.dat: %v\n", err)
		os.Exit(1)
	}

	err = copy.Copy(
		filepath.Join("desktop", "assets"),
		filepath.Join(outputDirectoryPath, "assets"),
	)
	if err != nil {
		fmt.Printf("Failed to copy desktop/assets: %v\n", err)
		os.Exit(1)
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
		fmt.Printf("Target platform %s is not supported, cgo_ldflags not implemented.\n", targetOS)
		os.Exit(1)
	}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working dir: %v\n", err)
		os.Exit(1)
	}

	cmdGoGetU := exec.Command(goBin, "get", "-u", "github.com/go-flutter-desktop/go-flutter")
	cmdGoGetU.Dir = filepath.Join(wd, "desktop")
	cmdGoGetU.Env = append(os.Environ(),
		"GO111MODULE=on",
		"CGO_LDFLAGS="+cgoLdflags,
	)
	cmdGoGetU.Stderr = os.Stderr
	cmdGoGetU.Stdout = os.Stdout

	err = cmdGoGetU.Run()
	if err != nil {
		fmt.Printf("Go get -u github.com/go-flutter-desktop/go-flutter failed: %v\n", err)
		os.Exit(1)
	}

	cmdGoModDownload := exec.Command(goBin, "mod", "download")
	cmdGoModDownload.Dir = filepath.Join(wd, "desktop")
	cmdGoModDownload.Env = append(os.Environ(),
		"GO111MODULE=on",
	)
	cmdGoModDownload.Stderr = os.Stderr
	cmdGoModDownload.Stdout = os.Stdout

	err = cmdGoModDownload.Run()
	if err != nil {
		fmt.Printf("Go mod download failed: %v\n", err)
		os.Exit(1)
	}

	cmdGoBuild := exec.Command(goBin, "build",
		"-o", outputBinaryPath,
		fmt.Sprintf("-ldflags=-X main.vmArguments=%s", strings.Join(vmArguments, ";")),
		dotSlash+"cmd",
	)
	cmdGoBuild.Dir = filepath.Join(wd, "desktop")
	cmdGoBuild.Env = append(os.Environ(),
		"GO111MODULE=on",
		"CGO_LDFLAGS="+cgoLdflags,
	)

	// set vars: (const?)
	// vmArguments

	cmdGoBuild.Stderr = os.Stderr
	cmdGoBuild.Stdout = os.Stdout

	err = cmdGoBuild.Run()
	if err != nil {
		fmt.Printf("Go build failed: %v\n", err)
		os.Exit(1)
	}
}
