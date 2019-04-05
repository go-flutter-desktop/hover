package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
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
		var pubspec struct {
			Name         string
			Dependencies map[string]interface{}
		}
		file, err := os.Open("pubspec.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Error: No pubspec.yaml file found.")
				goto Fail
			}
			fmt.Printf("Failed to open pubspec.yaml: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		err = yaml.NewDecoder(file).Decode(&pubspec)
		if err != nil {
			fmt.Printf("Failed to decode pubspec.yaml: %v\n", err)
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
