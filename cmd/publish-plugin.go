package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(publishPluginCmd)
}

var publishPluginCmd = &cobra.Command{
	Use:   "publish-plugin",
	Short: "Publish your go-flutter plugin as golang module in your github repo.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return errors.New("does not take arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterPluginProject()

		if build.GitBin == "" {
			log.Errorf("Failed to lookup `git` executable. Please install git")
			os.Exit(1)
		}

		// check if dir 'go' is tracked
		goCheckTrackedCmd := exec.Command(build.GitBin, "ls-files", "--error-unmatch", build.BuildPath)
		goCheckTrackedCmd.Stderr = os.Stderr
		err := goCheckTrackedCmd.Run()
		if err != nil {
			log.Errorf("The '%s' directory doesn't seems to be tracked by git. Error: %v", build.BuildPath, err)
			os.Exit(1)
		}

		// check if dir 'go' is clean (all tracked files are commited)
		goCheckCleanCmd := exec.Command(build.GitBin, "status", "--untracked-file=no", "--porcelain", build.BuildPath)
		goCheckCleanCmd.Stderr = os.Stderr
		cleanOut, err := goCheckCleanCmd.Output()
		if err != nil {
			log.Errorf("Failed to check if '%s' is clean.", build.BuildPath, err)
			os.Exit(1)
		}
		if len(cleanOut) != 0 {
			log.Errorf("The '%s' directory doesn't seems to be clean. (make sure tracked files are commited)", build.BuildPath)
			os.Exit(1)
		}

		// check if one of the git remote urls equals the package import 'url'
		pluginImportStr, err := readPluginGoImport(filepath.Join(build.BuildPath, "import.go.tmpl"), pubspec.GetPubSpec().Name)
		if err != nil {
			log.Errorf("Failed to read the plugin import url: %v", err)
			log.Infof("The file go/import.go.tmpl should look something like this:")
			fmt.Printf(`package main

import (
	flutter "github.com/go-flutter-desktop/go-flutter"
	%s "github.com/my-organization/%s/go"
)

// .. [init function] ..
      `, pubspec.GetPubSpec().Name, pubspec.GetPubSpec().Name)
			os.Exit(1)
		}
		url, err := url.Parse("https://" + pluginImportStr)
		if err != nil {
			log.Errorf("Failed to parse %s: %v", pluginImportStr, err)
			os.Exit(1)
		}
		// from go import string "github.com/my-organization/test_hover/go"
		// check if `git remote -v` has a match on:
		//  origin ?github.com?my-organization/test_hover.git
		// this regex works on https and ssh remotes.
		path := strings.TrimPrefix(url.Path, "/")
		path = strings.TrimSuffix(path, "/go")
		re := regexp.MustCompile(`(\w+)\s+(\S+)` + url.Host + "." + path + ".git")
		goCheckRemote := exec.Command(build.GitBin, "remote", "-v")
		goCheckRemote.Stderr = os.Stderr
		remoteOut, err := goCheckRemote.Output()
		if err != nil {
			log.Errorf("Failed to get git remotes: %v", err)
			os.Exit(1)
		}
		match := re.FindStringSubmatch(string(remoteOut))
		if len(match) < 1 {
			log.Warnf("At least one git remote urls must matchs the plugin golang import URL.")
			log.Printf("go import URL: %s", pluginImportStr)
			log.Printf("git remote -v:\n%s\n", string(remoteOut))
			goCheckRemote.Stdout = os.Stdout
			//default to origin
			log.Warnf("Assuming origin is where the plugin code is stored")
			log.Printf(" This warning can occur because the git repo name dosn't match the plugin name in pubspec.yaml")
			match = []string{"", "origin"}
		}

		tag := "go/v" + pubspec.GetPubSpec().Version

		log.Infof("Your plugin at version '%s' is ready to be publish as a golang module.", pubspec.GetPubSpec().Version)
		log.Infof("Please run: `%s`", log.Au().Magenta("git tag "+tag))
		log.Infof("            `%s`", log.Au().Magenta("git push "+match[1]+" "+tag))

		log.Infof(fmt.Sprintf("Let hover run those commands? "))
		if askForConfirmation() {
			gitTag := exec.Command(build.GitBin, "tag", tag)
			gitTag.Stderr = os.Stderr
			gitTag.Stdout = os.Stdout
			err = gitTag.Run()
			if err != nil {
				log.Errorf("The git command '%s' failed. Error: %v", gitTag.String(), err)
				os.Exit(1)
			}

			gitPush := exec.Command(build.GitBin, "push", match[1], tag)
			gitPush.Stderr = os.Stderr
			gitPush.Stdout = os.Stdout
			err = gitPush.Run()
			if err != nil {
				log.Errorf("The git command '%s' failed. Error: %v", gitPush.String(), err)
				os.Exit(1)
			}
		}

	},
}
