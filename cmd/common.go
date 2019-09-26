package cmd

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/go-flutter-desktop/hover/internal/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var (
	goBin      string
	flutterBin string
	dockerBin  string
)

// initBinaries is used to ensure go and flutter exec are found in the
// user's path
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

// PubSpec contains the parsed contents of pubspec.yaml
type PubSpec struct {
	Name         string
	Description  string
	Version      string
	Author       string
	Dependencies map[string]interface{}
	Flutter      map[string]interface{}
}

var pubspec = PubSpec{}

// getPubSpec returns the working directory pubspec.yaml as a PubSpec
func getPubSpec() PubSpec {
	if pubspec.Name == "" {
		pub, err := readPubSpecFile("pubspec.yaml")
		if err != nil {
			log.Errorf("%v", err)
			log.Errorf("This command should be run from the root of your Flutter project.")
			os.Exit(1)
		}
		pubspec = *pub
	}
	return pubspec
}

// readPubSpecFile reads a .yaml file at a path and return a correspond
// PubSpec struct
func readPubSpecFile(pubSpecPath string) (*PubSpec, error) {
	file, err := os.Open(pubSpecPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(err, "Error: No pubspec.yaml file found")
		}
		return nil, errors.Wrap(err, "Failed to open pubspec.yaml")
	}
	defer file.Close()

	var pub PubSpec
	err = yaml.NewDecoder(file).Decode(&pub)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode pubspec.yaml")
	}
	// avoid checking for the flutter dependencies for out of ws directories
	if pubSpecPath != "pubspec.yaml" {
		return &pub, nil
	}
	if _, exists := pub.Dependencies["flutter"]; !exists {
		return nil, errors.New(fmt.Sprintf("Missing `flutter` in %s dependencies list", pubSpecPath))
	}
	return &pub, nil
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

// hoverMigration migrates from old hover buildPath directory to the new one ("desktop" -> "go")
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
	fmt.Print(log.Au().Bold(log.Au().Cyan("hover: ")).String() + "[y/N]? ")
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

var camelcaseRegex = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")

// toCamelCase take a snake_case string and converts it to camelcase
func toCamelCase(str string) string {
	return camelcaseRegex.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}

// initializeGoModule uses the golang binary to initialize the go module
func initializeGoModule(projectPath string) {
	wd, err := os.Getwd()
	if err != nil {
		log.Errorf("Failed to get working dir: %v\n", err)
		os.Exit(1)
	}

	cmdGoModInit := exec.Command(goBin, "mod", "init", projectPath+"/"+buildPath)
	cmdGoModInit.Dir = filepath.Join(wd, buildPath)
	cmdGoModInit.Env = append(os.Environ(),
		"GO111MODULE=on",
	)
	cmdGoModInit.Stderr = os.Stderr
	cmdGoModInit.Stdout = os.Stdout
	err = cmdGoModInit.Run()
	if err != nil {
		log.Errorf("Go mod init failed: %v\n", err)
		os.Exit(1)
	}

	cmdGoModTidy := exec.Command(goBin, "mod", "tidy")
	cmdGoModTidy.Dir = filepath.Join(wd, buildPath)
	log.Infof("go-flutter project is located: " + cmdGoModTidy.Dir)
	cmdGoModTidy.Env = append(os.Environ(),
		"GO111MODULE=on",
	)
	cmdGoModTidy.Stderr = os.Stderr
	cmdGoModTidy.Stdout = os.Stdout
	err = cmdGoModTidy.Run()
	if err != nil {
		log.Errorf("Go mod tidy failed: %v\n", err)
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

// shouldRunPluginGet checks if the pubspec.yaml file is older than the
// .packages file, if it is the case, prompt the user for a hover plugin get.
func shouldRunPluginGet() (bool, error) {
	file1Info, err := os.Stat("pubspec.yaml")
	if err != nil {
		return false, err
	}

	file2Info, err := os.Stat(".packages")
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	modTime1 := file1Info.ModTime()
	modTime2 := file2Info.ModTime()

	diff := modTime1.Sub(modTime2)

	if diff > (time.Duration(0) * time.Second) {
		return true, nil
	}
	return false, nil
}
