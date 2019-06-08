package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/go-flutter-desktop/hover/internal/pubspec"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	cobra.OnInitialize(initBinaries)
}

var (
	goBin      string
	flutterBin string
)

func initBinaries() {
	var err error
	goBin, err = exec.LookPath("go")
	if err != nil {
		fmt.Println("Failed to lookup `go` executable. Please install Go.\nhttps://golang.org/doc/install")
		os.Exit(1)
	}
	flutterBin, err = exec.LookPath("flutter")
	if err != nil {
		fmt.Println("Failed to lookup `flutter` executable. Please install flutter.\nhttps://flutter.dev/docs/get-started/install")
		os.Exit(1)
	}
}

// assertInFlutterProject asserts this command is executed in a flutter project
// and returns the project name.
func assertInFlutterProject() string {
	{
		pubspec, err := pubspec.ReadLocal()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			goto Fail
		}
		if _, exists := pubspec.Dependencies["flutter"]; !exists {
			fmt.Println("Missing `flutter` in pubspec.yaml dependencies list.")
			goto Fail
		}

		return pubspec.Name
	}

Fail:
	fmt.Println("This command should be run from the root of your Flutter project.")
	os.Exit(1)
	return ""
}

func assertHoverInitialized() {
	_, err := os.Stat("desktop")
	if os.IsNotExist(err) {
		fmt.Println("Directory 'desktop' is missing, did you run `hover init` in this project?")
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("Failed to detect directory desktop: %v\n", err)
		os.Exit(1)
	}
}

// findPubcachePath returns the absolute path for the pub-cache or an error.
func findPubcachePath() (string, error) {
	var path string
	switch runtime.GOOS {
	case "darwin", "linux":
		home, err := homedir.Dir()
		if err != nil {
			return "", errors.Wrap(err, "failed to resolve user home dir")
		}
		path = filepath.Join(home, ".pub-cache")
	case "windows":
		path = filepath.Join(os.Getenv("APPDATA"), "Pub", "Cache")
	}
	return path, nil
}
