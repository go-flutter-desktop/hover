package cmd

import (
	"bufio"
	"encoding/xml"
	"github.com/go-flutter-desktop/hover/internal/log"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-flutter-desktop/hover/internal/pubspec"
	"github.com/go-flutter-desktop/hover/internal/build"
)

func initBinaries() {
	var err error
	goAvailable := false
	dockerAvailable := false
	build.GoBin, err = exec.LookPath("go")
	if err == nil {
		goAvailable = true
	}
	build.DockerBin, err = exec.LookPath("docker")
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
	build.FlutterBin, err = exec.LookPath("flutter")
	if err != nil {
		log.Errorf("Failed to lookup 'flutter' executable. Please install flutter.\nhttps://flutter.dev/docs/get-started/install")
		os.Exit(1)
	}
}

// assertInFlutterProject asserts this command is executed in a flutter project
func assertInFlutterProject() {
	pubspec.GetPubSpec()
}

func assertHoverInitialized() {
	_, err := os.Stat(build.BuildPath)
	if os.IsNotExist(err) {
		if hoverMigration() {
			return
		}
		log.Errorf("Directory '%s' is missing. Please init go-flutter first: %s", build.BuildPath, log.Au().Magenta("hover init"))
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
		err := os.Rename(oldBuildPath, build.BuildPath)
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
