package cmd

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-flutter-desktop/hover/internal/versioncheck"
	version "github.com/hashicorp/go-version"
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
		fmt.Println("hover: Failed to lookup `go` executable. Please install Go.\nhttps://golang.org/doc/install")
		os.Exit(1)
	}
	flutterBin, err = exec.LookPath("flutter")
	if err != nil {
		fmt.Println("hover: Failed to lookup `flutter` executable. Please install flutter.\nhttps://flutter.dev/docs/get-started/install")
		os.Exit(1)
	}
}

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
					fmt.Println("hover: Error: No pubspec.yaml file found.")
					goto Fail
				}
				fmt.Printf("hover: Failed to open pubspec.yaml: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()

			err = yaml.NewDecoder(file).Decode(&pubspec)
			if err != nil {
				fmt.Printf("hover: Failed to decode pubspec.yaml: %v\n", err)
				goto Fail
			}
			if _, exists := pubspec.Dependencies["flutter"]; !exists {
				fmt.Println("hover: Missing `flutter` in pubspec.yaml dependencies list.")
				goto Fail
			}
		}

		return pubspec
	}

Fail:
	fmt.Println("hover: This command should be run from the root of your Flutter project.")
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
		fmt.Println("hover: Directory '" + buildPath + "' is missing, did you run `hover init` in this project?")
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("hover: Failed to detect directory desktop: %v\n", err)
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

	fmt.Println("hover: ⚠ Found older hover directory layout, hover is now expecting a 'go' directory instead of 'desktop'.")
	fmt.Println("hover: ⚠    To migrate, rename the 'desktop' directory to 'go'.")
	fmt.Printf("hover:      Let hover do the migration? ")

	if askForConfirmation() {
		err := os.Rename(oldBuildPath, buildPath)
		if err != nil {
			fmt.Printf("hover: Migration failed: %v\n", err)
			return false
		}
		fmt.Printf("hover: Migration success\n")
		return true
	}

	return false
}

// askForConfirmation asks the user for confirmation.
func askForConfirmation() bool {
	fmt.Printf("[y/N]: ")
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

// AndroidManifest is a file that describes the essential information about
// an android app.
type AndroidManifest struct {
	Package string `xml:"package,attr"`
}

func addBuildConstantSourceFile() {
	currentEmbedderTag, err := versioncheck.CurrentGoFlutterTag(buildPath)
	if err != nil {
		fmt.Printf("hover: %v\n", err)
		os.Exit(1)
	}

	semver, err := version.NewSemver(currentEmbedderTag)
	if err != nil {
		fmt.Printf("hover: faild to parse 'go-flutter' semver: %v\n", err)
		os.Exit(1)
	}

	minimumGoFlutterRelease, _ := version.NewSemver("0.30.0")
	if semver.LessThan(minimumGoFlutterRelease) {
		return
	}

	buildVersionPath := filepath.Join(buildPath, "cmd", "version.go")
	if _, err := os.Stat(buildVersionPath); err == nil {
		// file exist
		err := os.Remove(buildVersionPath)
		if err != nil {
			fmt.Printf("hover: Failed to delete %s: %v\n", buildVersionPath, err)
			os.Exit(1)
		}
	}

	buildVersionFile, err := os.Create(buildVersionPath)
	if err != nil {
		fmt.Printf("hover: Failed to create %s: %v\n", buildVersionPath, err)
		os.Exit(1)
	}

	// Default value
	androidManifestFile := "android/app/src/main/AndroidManifest.xml"

	// Open AndroidManifest file
	xmlFile, err := os.Open(androidManifestFile)
	if err != nil {
		fmt.Printf("hover: Failed to retrieve the organization name: %v\n", err)
		writeVersionFle(currentEmbedderTag, "com.example", buildVersionFile)
		return
	}
	defer xmlFile.Close()

	byteXMLValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		fmt.Printf("hover: Failed to retrieve the organization name: %v\n", err)
		writeVersionFle(currentEmbedderTag, "com.example", buildVersionFile)
		return
	}

	var androidManifest AndroidManifest
	err = xml.Unmarshal(byteXMLValue, &androidManifest)
	if err != nil {
		fmt.Printf("hover: Failed to retrieve the organization name: %v\n", err)
		writeVersionFle(currentEmbedderTag, "com.example", buildVersionFile)
		return
	}
	javaPackage := strings.Split(androidManifest.Package, ".")
	orgName := strings.Join(javaPackage[:len(javaPackage)-1], ".")
	if orgName == "" {
		writeVersionFle(currentEmbedderTag, "com.example", buildVersionFile)
		return
	}

	writeVersionFle(currentEmbedderTag, orgName, buildVersionFile)
}

func writeVersionFle(currentEmbedderTag string, organizationName string, buildVersionFile *os.File) {
	buildVersionFileContent := []string{
		"package main",
		"",
		"import (",
		"	\"github.com/go-flutter-desktop/go-flutter\"",
		")",
		"",
		"func init() {",
		"	// DO NOT EDIT, may be set by hover at compile-time",
		"	flutter.SetPlatformVersion(\"" + currentEmbedderTag + "\")",
		"	flutter.SetProjectVersion(\"" + getPubSpec().Version + "\")",
		"	flutter.SetAppName(\"" + getPubSpec().Name + "\")",
		"	flutter.SetOrganizationName(\"" + organizationName + "\")",
		"}",
	}

	for _, line := range buildVersionFileContent {
		if _, err := buildVersionFile.WriteString(line + "\n"); err != nil {
			fmt.Printf("hover: Could not write version.go file: %v\n", err)
			os.Exit(1)
		}
	}
	err := buildVersionFile.Close()
	if err != nil {
		fmt.Printf("hover: Could not close version.go file: %v\n", err)
		os.Exit(1)
	}
}
