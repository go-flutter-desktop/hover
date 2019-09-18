package cmd

import (
	"bufio"
	"github.com/go-flutter-desktop/hover/internal/log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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
		log.Fatal("Failed to lookup `go` executable. Please install Go.\nhttps://golang.org/doc/install")
		os.Exit(1)
	}
	flutterBin, err = exec.LookPath("flutter")
	if err != nil {
		log.Fatal("Failed to lookup `flutter` executable. Please install flutter.\nhttps://flutter.dev/docs/get-started/install")
		os.Exit(1)
	}
}

// PubSpec  basic model pubspec
type PubSpec struct {
	Name         string
	Description  string
	Version      string
	Author       string
	Dependencies map[string]interface{}
}

var pubspec = PubSpec{}

func getPubSpec() PubSpec {
	{
		if pubspec.Name == "" {
			file, err := os.Open("pubspec.yaml")
			if err != nil {
				if os.IsNotExist(err) {
					log.Fatal("Error: No pubspec.yaml file found.")
					goto Fail
				}
				log.Fatal("Failed to open pubspec.yaml: %v", err)
				os.Exit(1)
			}
			defer file.Close()

			err = yaml.NewDecoder(file).Decode(&pubspec)
			if err != nil {
				log.Fatal("Failed to decode pubspec.yaml: %v", err)
				goto Fail
			}
			if _, exists := pubspec.Dependencies["flutter"]; !exists {
				log.Fatal("Missing `flutter` in pubspec.yaml dependencies list.")
				goto Fail
			}
		}

		return pubspec
	}

Fail:
	log.Fatal("This command should be run from the root of your Flutter project.")
	os.Exit(1)
	return PubSpec{}
}

// assertInFlutterProject asserts this command is executed in a flutter project
func assertInFlutterProject() {
	getPubSpec()
}

func assertHoverInitialized() {
	_, err := os.Stat(buildPath)
	if os.IsNotExist(err) {
		if hoverMigration() {
			return
		}
		log.Fatal("Directory '" + buildPath + "' is missing, did you run `hover init` in this project?")
		os.Exit(1)
	}
	if err != nil {
		log.Fatal("Failed to detect directory desktop: %v", err)
		os.Exit(1)
	}
}

func hoverMigration() bool {
	oldBuildPath := "desktop"
	file, err := os.Open(filepath.Join(oldBuildPath, "go.mod"))
	if err != nil {
		return false
	}
	defer file.Close()

	log.Warn("⚠ Found older hover directory layout, hover is now expecting a 'go' directory instead of 'desktop'.")
	log.Warn("⚠    To migrate, rename the 'desktop' directory to 'go'.")
	log.Warn("     Let hover do the migration? ")

	if askForConfirmation() {
		err := os.Rename(oldBuildPath, buildPath)
		if err != nil {
			log.Warn("Migration failed: %v", err)
			return false
		}
		log.Info("Migration success")
		return true
	}

	return false
}

// askForConfirmation asks the user for confirmation.
func askForConfirmation() bool {
	log.Print("[y/N]: ")
	in := bufio.NewReader(os.Stdin)
	s, err := in.ReadString('\n')
	if err != nil {
		panic(err)
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return false
}
