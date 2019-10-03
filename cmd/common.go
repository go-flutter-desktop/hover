package cmd

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

// initBinaries is used to ensure go and flutter exec are found in the
// user's path
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
	gitBin, err = exec.LookPath("git")
	if err != nil {
		log.Warnf("Failed to lookup 'git' executable.")
	}
}

// assertInFlutterProject asserts this command is executed in a flutter project
func assertInFlutterProject() {
	pubspec.GetPubSpec()
}

// assertInFlutterPluginProject asserts this command is executed in a flutter plugin project
func assertInFlutterPluginProject() {
	if _, ok := pubspec.GetPubSpec().Flutter["plugin"]; !ok {
		log.Errorf("The directory doesn't appear to contain a plugin package.\nTo create a new plugin, first run `%s`, then run `%s`.", log.Au().Magenta("flutter create --template=plugin"), log.Au().Magenta("hover init-plugin"))
		os.Exit(1)
	}
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
