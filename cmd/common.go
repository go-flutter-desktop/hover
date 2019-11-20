package cmd

import (
	"bufio"
	"fmt"
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
	"github.com/pkg/errors"
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
		log.Errorf("Failed to lookup `go` executable. Please install go or add `--docker` to force running in Docker container.\nhttps://golang.org/doc/install")
		os.Exit(1)
	}
	build.FlutterBin, err = exec.LookPath("flutter")
	if err != nil {
		log.Errorf("Failed to lookup `flutter` executable. Please install flutter.\nhttps://flutter.dev/docs/get-started/install")
		os.Exit(1)
	}
	build.GitBin, err = exec.LookPath("git")
	if err != nil {
		log.Warnf("Failed to lookup `git` executable.")
	}
}

// assertInFlutterProject asserts this command is executed in a flutter project
func assertInFlutterProject() {
	pubspec.GetPubSpec()
}

// assertInFlutterPluginProject asserts this command is executed in a flutter plugin project
func assertInFlutterPluginProject() {
	if _, ok := pubspec.GetPubSpec().Flutter["plugin"]; !ok {
		log.Errorf("The directory doesn`t appear to contain a plugin package.\nTo create a new plugin, first run `%s`, then run `%s`.", log.Au().Magenta("flutter create --template=plugin"), log.Au().Magenta("hover init-plugin"))
		os.Exit(1)
	}
}

func assertHoverInitialized() {
	_, err := os.Stat(build.BuildPath)
	if os.IsNotExist(err) {
		if hoverMigration() {
			return
		}
		log.Errorf("Directory `%s` is missing. Please init go-flutter first: %s", build.BuildPath, log.Au().Magenta("hover init"))
		os.Exit(1)
	}
	if err != nil {
		log.Errorf("Failed to detect directory desktop: %v", err)
		os.Exit(1)
	}
}

func checkFlutterChannel() {
	cmdCheckFlutter := exec.Command(build.FlutterBin, "--version")
	cmdCheckFlutterOut, err := cmdCheckFlutter.Output()
	if err != nil {
		log.Warnf("Failed to check your flutter channel: %v", err)
	} else {
		re := regexp.MustCompile("•\\schannel\\s(\\w*)\\s•")

		match := re.FindStringSubmatch(string(cmdCheckFlutterOut))
		if len(match) >= 2 {
			ignoreWarning := os.Getenv("HOVER_IGNORE_CHANNEL_WARNING")
			if match[1] != "beta" && ignoreWarning != "true" {
				log.Warnf("⚠ The go-flutter project tries to stay compatible with the beta channel of Flutter.")
				log.Warnf("⚠     It's advised to use the beta channel: `%s`", log.Au().Magenta("flutter channel beta"))
			}
		} else {
			log.Warnf("Failed to check your flutter channel: Unrecognized output format")
		}
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

	log.Warnf("⚠ Found older hover directory layout, hover is now expecting a `go` directory instead of `desktop`.")
	log.Warnf("⚠    To migrate, rename the `desktop` directory to `go`.")
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
		log.Errorf("Failed to get working dir: %v", err)
		os.Exit(1)
	}

	cmdGoModInit := exec.Command(build.GoBin, "mod", "init", projectPath+"/"+build.BuildPath)
	cmdGoModInit.Dir = filepath.Join(wd, build.BuildPath)
	cmdGoModInit.Env = append(os.Environ(),
		"GO111MODULE=on",
	)
	cmdGoModInit.Stderr = os.Stderr
	cmdGoModInit.Stdout = os.Stdout
	err = cmdGoModInit.Run()
	if err != nil {
		log.Errorf("Go mod init failed: %v", err)
		os.Exit(1)
	}

	cmdGoModTidy := exec.Command(build.GoBin, "mod", "tidy")
	cmdGoModTidy.Dir = filepath.Join(wd, build.BuildPath)
	log.Infof("You can add the `%s` directory to git.", cmdGoModTidy.Dir)
	cmdGoModTidy.Env = append(os.Environ(),
		"GO111MODULE=on",
	)
	cmdGoModTidy.Stderr = os.Stderr
	cmdGoModTidy.Stdout = os.Stdout
	err = cmdGoModTidy.Run()
	if err != nil {
		log.Errorf("Go mod tidy failed: %v", err)
		os.Exit(1)
	}
}

// findPubcachePath returns the absolute path for the pub-cache or an error.
func findPubcachePath() (string, error) {
	var path string
	switch runtime.GOOS {
	case "darwin", "linux":
		home, err := os.UserHomeDir()
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
