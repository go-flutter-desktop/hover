package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/otiai10/copy"
	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/internal/enginecache"
)

var dotSlash = string([]byte{'.', filepath.Separator})

var targetMainDart string

func init() {
	buildCmd.Flags().StringVarP(&targetMainDart, "target", "t", "lib/main.dart", "The main entry-point file of the application.")
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a desktop release",
	Run: func(cmd *cobra.Command, args []string) {

		projectName := assertInFlutterProject()

		// Hardcode target to the current OS (no cross-compile support yet)
		targetOS := runtime.GOOS

		outputDirectoryPath, err := filepath.Abs(filepath.Join("desktop", "build", "outputs", targetOS))
		if err != nil {
			fmt.Printf("Failed to resolve absolute path for output directory: %v\n", err)
			os.Exit(1)
		}

		var outputBinaryPath = filepath.Join(outputDirectoryPath, projectName)
		switch targetOS {
		case "linux":
			// no special filename
		case "darwin":
			// no special filename
		case "windows":
			outputBinaryPath += ".exe"
		default:
			fmt.Printf("Target platform %s is not supported.\n", targetOS)
			os.Exit(1)
		}

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
			"--target", targetMainDart,
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

		cmdGoModDownload := exec.Command(goBin, "mod", "download")
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Failed to get working dir: %v\n", err)
			os.Exit(1)
		}
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
			dotSlash+"cmd",
		)
		cmdGoBuild.Dir = filepath.Join(wd, "desktop")
		cmdGoBuild.Env = append(os.Environ(),
			"GO111MODULE=on",
			"CGO_LDFLAGS="+cgoLdflags,
		)

		// static build

		// set vars: (const?)
		// vmArguments

		cmdGoBuild.Stderr = os.Stderr
		cmdGoBuild.Stdout = os.Stdout

		err = cmdGoBuild.Run()
		if err != nil {
			fmt.Printf("Go build failed: %v\n", err)
			os.Exit(1)
		}
	},
}
