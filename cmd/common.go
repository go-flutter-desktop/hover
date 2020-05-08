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

	"github.com/pkg/errors"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/flutterversion"
	"github.com/go-flutter-desktop/hover/internal/logx"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

// assertInFlutterProject asserts this command is executed in a flutter project
func assertInFlutterProject() {
	pubspec.GetPubSpec()
}

// assertInFlutterPluginProject asserts this command is executed in a flutter plugin project
func assertInFlutterPluginProject() {
	if _, ok := pubspec.GetPubSpec().Flutter["plugin"]; !ok {
		logx.Errorf("The directory doesn't appear to contain a plugin package.\nTo create a new plugin, first run `%s`, then run `%s`.", logx.Au().Magenta("flutter create --template=plugin"), logx.Au().Magenta("hover init-plugin"))
		os.Exit(1)
	}
}

func assertHoverInitialized() {
	_, err := os.Stat(build.BuildPath)
	if os.IsNotExist(err) {
		if hoverMigrateDesktopToGo() {
			return
		}
		logx.Errorf("Directory '%s' is missing. Please init go-flutter first: %s", build.BuildPath, logx.Au().Magenta("hover init"))
		os.Exit(1)
	}
	if err != nil {
		logx.Errorf("Failed to detect directory desktop: %v", err)
		os.Exit(1)
	}
}

func checkFlutterChannel() {
	channel := flutterversion.FlutterChannel()
	ignoreWarning := os.Getenv("HOVER_IGNORE_CHANNEL_WARNING")
	if channel != "beta" && ignoreWarning != "true" {
		logx.Warnf("⚠ The go-flutter project tries to stay compatible with the beta channel of Flutter.")
		logx.Warnf("⚠     It's advised to use the beta channel: `%s`", logx.Au().Magenta("flutter channel beta"))
	}
}

// hoverMigrateDesktopToGo migrates from old hover buildPath directory to the new one ("desktop" -> "go")
func hoverMigrateDesktopToGo() bool {
	oldBuildPath := "desktop"
	file, err := os.Open(filepath.Join(oldBuildPath, "go.mod"))
	if err != nil {
		return false
	}
	defer file.Close()

	logx.Warnf("⚠ Found older hover directory layout, hover is now expecting a 'go' directory instead of 'desktop'.")
	logx.Warnf("⚠    To migrate, rename the 'desktop' directory to 'go'.")
	logx.Warnf("     Let hover do the migration? ")

	if askForConfirmation() {
		err := os.Rename(oldBuildPath, build.BuildPath)
		if err != nil {
			logx.Warnf("Migration failed: %v", err)
			return false
		}
		logx.Infof("Migration success")
		return true
	}

	return false
}

// askForConfirmation asks the user for confirmation.
func askForConfirmation() bool {
	fmt.Print(logx.Au().Bold(logx.Au().Cyan("hover: ")).String() + "[y/N]? ")

	if len(os.Getenv("HOVER_DISABLE_INTERACTIONS")) > 0 {
		fmt.Println(logx.Au().Bold(logx.Au().Yellow("Interactions disabled, assuming 'no'.")).String())
		return false
	}

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
		logx.Errorf("Failed to get working dir: %v\n", err)
		os.Exit(1)
	}

	cmdGoModInit := exec.Command(build.GoBin(), "mod", "init", projectPath+"/"+build.BuildPath)
	cmdGoModInit.Dir = filepath.Join(wd, build.BuildPath)
	cmdGoModInit.Env = append(os.Environ(),
		"GO111MODULE=on",
	)
	cmdGoModInit.Stderr = os.Stderr
	cmdGoModInit.Stdout = os.Stdout
	err = cmdGoModInit.Run()
	if err != nil {
		logx.Errorf("Go mod init failed: %v\n", err)
		os.Exit(1)
	}

	cmdGoModTidy := exec.Command(build.GoBin(), "mod", "tidy")
	cmdGoModTidy.Dir = filepath.Join(wd, build.BuildPath)
	logx.Infof("You can add the '%s' directory to git.", cmdGoModTidy.Dir)
	cmdGoModTidy.Env = append(os.Environ(),
		"GO111MODULE=on",
	)
	cmdGoModTidy.Stderr = os.Stderr
	cmdGoModTidy.Stdout = os.Stdout
	err = cmdGoModTidy.Run()
	if err != nil {
		logx.Errorf("Go mod tidy failed: %v\n", err)
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
