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

	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(publishPluginCmd)
}

var publishPluginCmd = &cobra.Command{
	Use:   "publish-plugin",
	Short: "Assert that your go-flutter plugin can be pushed as golang module in your github repo.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return errors.New("does not take arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterPluginProject()

		if goBin == "" {
			log.Errorf("Failed to lookup `git` executable. Please install git")
			os.Exit(1)
		}

		// check if dir 'go' is tracked
		goCheckTrackedCmd := exec.Command(gitBin, "ls-files", "--error-unmatch", buildPath)
		goCheckTrackedCmd.Stderr = os.Stderr
		err := goCheckTrackedCmd.Run()
		if err != nil {
			log.Errorf("The '%s' directory doesn't seems to be tracked by git. Error: %v", buildPath, err)
			os.Exit(1)
		}

		// check if dir 'go' is clean (all tracked files are commited)
		goCheckCleanCmd := exec.Command(gitBin, "status", "--untracked-file=no", "--porcelain", buildPath)
		goCheckCleanCmd.Stderr = os.Stderr
		cleanOut, err := goCheckCleanCmd.Output()
		if err != nil {
			log.Errorf("Failed to check if '%s' is clean.", buildPath, err)
			os.Exit(1)
		}
		if len(cleanOut) != 0 {
			log.Errorf("The '%s' directory doesn't seems to be clean. (make sure tracked files are commited)", buildPath)
			os.Exit(1)
		}

		// check if one of the git remote urls equals the package import 'url'
		pluginImportStr, err := readPluginGoImport(filepath.Join("go", "import.go.tmpl"), getPubSpec().Name)
		if err != nil {
			log.Errorf("Failed to read the plugin import url: %v", err)
			log.Infof("The file go/import.go.tmpl should look something like this:")
			fmt.Printf(`package main

import (
	flutter "github.com/go-flutter-desktop/go-flutter"
	%s "github.com/my-organization/%s/go"
)

// .. [init function] ..
      `, getPubSpec().Name, getPubSpec().Name)
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
		goCheckRemote := exec.Command(gitBin, "remote", "-v")
		goCheckRemote.Stderr = os.Stderr
		remoteOut, err := goCheckRemote.Output()
		if err != nil {
			log.Errorf("Failed to get git remotes: %v", err)
			os.Exit(1)
		}
		match := re.FindStringSubmatch(string(remoteOut))
		if len(match) < 1 {
			log.Errorf("At least one git remote urls must matchs the plugin golang import URL.")
			log.Printf("go import URL: %s", pluginImportStr)
			log.Printf("git remote -v:\n%s\n", string(remoteOut))
			goCheckRemote.Stdout = os.Stdout
			os.Exit(1)
		}

		tag := "go/v" + getPubSpec().Version

		log.Infof("Your plugin at version '%s' is ready to be publish as a golang module.", getPubSpec().Version)
		log.Infof("Please run: `%s`", log.Au().Magenta("git tag "+tag))
		log.Infof("            `%s`", log.Au().Magenta("git push "+match[1]+" "+tag))

		log.Infof(fmt.Sprintf("Let hover run those commands? "))
		if askForConfirmation() {
			gitTag := exec.Command(gitBin, "tag", tag)
			gitTag.Stderr = os.Stderr
			gitTag.Stdout = os.Stdout
			err = gitTag.Run()
			if err != nil {
				log.Errorf("The git command '%s' failed. Error: %v", gitTag.String(), err)
				os.Exit(1)
			}

			gitPush := exec.Command(gitBin, "push", match[1], tag)
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
