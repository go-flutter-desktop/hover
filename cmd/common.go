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

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	cobra.OnInitialize(initBinaries)
}

var (
	goBin      string
	flutterBin string
	dockerBin  string
)

func initBinaries() {
	var err error
	goAvailable := false
	dockerAvailable := false
	goBin, err = exec.LookPath("go")
	if err == nil {
		goAvailable = true
	}
	dockerBin, err = exec.LookPath("docker")
	if err == nil {
		dockerAvailable = true
	}
	if !dockerAvailable && !goAvailable {
		fmt.Println("hover: Failed to lookup `go` and `docker` executable. Please install one of them:\nGo: https://golang.org/doc/install\nDocker: https://docs.docker.com/install")
		os.Exit(1)
	}
	if dockerAvailable && !goAvailable && !buildDocker {
		fmt.Println("hover: Failed to lookup `go` executable. Please install go or add '--docker' to force running in Docker container.\nhttps://golang.org/doc/install")
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

// androidOrganizationName fetch the android package name (default:
// 'com.example').
// Can by set upon flutter create (--org flag)
//
// If errors occurs when reading the android package name, the string value
// will correspond to 'build.hover.example'
func androidOrganizationName() string {
	// Default value
	androidManifestFile := "android/app/src/main/AndroidManifest.xml"

	// Open AndroidManifest file
	xmlFile, err := os.Open(androidManifestFile)
	if err != nil {
		fmt.Printf("hover: Failed to retrieve the organization name: %v\n", err)
		return "build.hover.example"
	}
	defer xmlFile.Close()

	byteXMLValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		fmt.Printf("hover: Failed to retrieve the organization name: %v\n", err)
		return "build.hover.example"
	}

	var androidManifest AndroidManifest
	err = xml.Unmarshal(byteXMLValue, &androidManifest)
	if err != nil {
		fmt.Printf("hover: Failed to retrieve the organization name: %v\n", err)
		return "build.hover.example"
	}
	javaPackage := strings.Split(androidManifest.Package, ".")
	orgName := strings.Join(javaPackage[:len(javaPackage)-1], ".")
	if orgName == "" {
		return "build.hover.example"
	}
	return orgName
}
