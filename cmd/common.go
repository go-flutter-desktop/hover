package cmd

import (
	"bufio"
	"encoding/xml"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-flutter-desktop/hover/internal/log"
	"gopkg.in/yaml.v2"
)

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
		log.Errorf("Failed to lookup `go` and `docker` executable. Please install one of them:\nGo: https://golang.org/doc/install\nDocker: https://docs.docker.com/install")
		os.Exit(1)
	}
	if dockerAvailable && !goAvailable && !buildDocker {
		log.Errorf("Failed to lookup `go` executable. Please install go or add '--docker' to force running in Docker container.\nhttps://golang.org/doc/install")
		os.Exit(1)
	}
	flutterBin, err = exec.LookPath("flutter")
	if err != nil {
		log.Errorf("Failed to lookup 'flutter' executable. Please install flutter.\nhttps://flutter.dev/docs/get-started/install")
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
					log.Errorf("Error: No pubspec.yaml file found.")
					goto Fail
				}
				log.Errorf("Failed to open pubspec.yaml: %v", err)
				os.Exit(1)
			}
			defer file.Close()

			err = yaml.NewDecoder(file).Decode(&pubspec)
			if err != nil {
				log.Errorf("Failed to decode pubspec.yaml: %v", err)
				goto Fail
			}
			if _, exists := pubspec.Dependencies["flutter"]; !exists {
				log.Errorf("Missing 'flutter' in pubspec.yaml dependencies list.")
				goto Fail
			}
		}

		return pubspec
	}

Fail:
	log.Errorf("This command should be run from the root of your Flutter project.")
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
		log.Errorf("Directory '%s' is missing. Please init go-flutter first: %s", buildPath, log.Au().Magenta("hover init"))
		os.Exit(1)
	}
	if err != nil {
		log.Errorf("Failed to detect directory desktop: %v", err)
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

	log.Warnf("⚠ Found older hover directory layout, hover is now expecting a 'go' directory instead of 'desktop'.")
	log.Warnf("⚠    To migrate, rename the 'desktop' directory to 'go'.")
	log.Warnf("     Let hover do the migration? ")

	if askForConfirmation() {
		err := os.Rename(oldBuildPath, buildPath)
		if err != nil {
			log.Warnf("Migration failed: %v", err)
			return false
		}
		log.Infof("Migration success")
		return true
	}

	return false
}

// askForConfirmation asks the user for confirmation.
func askForConfirmation() bool {
	log.Printf("[y/N]: ")
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
// will correspond to 'hover.failed.to.retrieve.package.name'
func androidOrganizationName() string {
	// Default value
	androidManifestFile := "android/app/src/main/AndroidManifest.xml"

	// Open AndroidManifest file
	xmlFile, err := os.Open(androidManifestFile)
	if err != nil {
		log.Errorf("Failed to retrieve the organization name: %v", err)
		return "hover.failed.to.retrieve.package.name"
	}
	defer xmlFile.Close()

	byteXMLValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Errorf("Failed to retrieve the organization name: %v", err)
		return "hover.failed.to.retrieve.package.name"
	}

	var androidManifest AndroidManifest
	err = xml.Unmarshal(byteXMLValue, &androidManifest)
	if err != nil {
		log.Errorf("Failed to retrieve the organization name: %v", err)
		return "hover.failed.to.retrieve.package.name"
	}
	javaPackage := strings.Split(androidManifest.Package, ".")
	orgName := strings.Join(javaPackage[:len(javaPackage)-1], ".")
	if orgName == "" {
		return "hover.failed.to.retrieve.package.name"
	}
	return orgName
}
