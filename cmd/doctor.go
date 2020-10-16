package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/go-flutter-desktop/hover/cmd/packaging"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/config"
	"github.com/go-flutter-desktop/hover/internal/flutterversion"
	"github.com/go-flutter-desktop/hover/internal/log"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Show information about the installed tooling",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("does not take arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterProject()

		version := hoverVersion()
		log.Infof("Hover version %s running on %s", version, runtime.GOOS)

		log.Infof("Sharing packaging tools")
		for _, task := range []packaging.Task{
			packaging.DarwinBundleTask,
			packaging.DarwinDmgTask,
			packaging.DarwinPkgTask,
			packaging.LinuxAppImageTask,
			packaging.LinuxDebTask,
			packaging.LinuxPkgTask,
			packaging.LinuxRpmTask,
			packaging.LinuxSnapTask,
			packaging.WindowsMsiTask,
		} {
			if task.IsSupported() {
				log.Infof("%s is supported", task.Name())
			}
		}
		log.Printf("")

		log.Infof("Sharing flutter version")
		cmdFlutterVersion := exec.Command(build.FlutterBin(), "--version")
		cmdFlutterVersion.Stderr = os.Stderr
		cmdFlutterVersion.Stdout = os.Stdout
		err := cmdFlutterVersion.Run()
		if err != nil {
			log.Errorf("Flutter --version failed: %v", err)
		}

		engineCommitHash := flutterversion.FlutterRequiredEngineVersion()
		log.Infof("Flutter engine commit: %s", log.Au().Magenta("https://github.com/flutter/engine/commit/"+engineCommitHash))

		checkFlutterChannel()

		cmdGoEnvCC := exec.Command(build.GoBin(), "env", "CC")
		cmdGoEnvCCOut, err := cmdGoEnvCC.Output()
		if err != nil {
			log.Errorf("Go env CC failed: %v", err)
		}
		cCompiler := strings.Trim(string(cmdGoEnvCCOut), " ")
		cCompiler = strings.Trim(cCompiler, "\n")
		if cCompiler != "" {
			log.Infof("Finding out the C compiler version")
			cmdCCVersion := exec.Command(cCompiler, "--version")
			cmdCCVersion.Stderr = os.Stderr
			cmdCCVersion.Stdout = os.Stdout
			cmdCCVersion.Run()
		}

		log.Infof("Sharing the content of go.mod")
		file, err := os.Open(filepath.Join(build.BuildPath, "go.mod"))
		if err != nil {
			log.Errorf("Failed to read go.mod: %v", err)
		} else {
			defer file.Close()
			b, _ := ioutil.ReadAll(file)
			fmt.Print(string(b))
		}

		hoverConfig, err := config.ReadConfigFile(filepath.Join(build.BuildPath, "hover.yaml"))
		if err != nil {
			log.Warnf("%v", err)
		} else {
			log.Infof("Sharing the content of hover.yaml")
			dump, err := yaml.Marshal(hoverConfig)
			if err != nil {
				log.Warnf("%v", err)
			} else {
				fmt.Print(string(dump))
			}
		}

		log.Infof("Sharing the content of go/cmd")
		files, err := filepath.Glob(filepath.Join(build.BuildPath, "cmd", "*"))
		if err != nil {
			log.Errorf("Failed to get the list of files in go/cmd", err)
			os.Exit(1)
		}
		fmt.Println(strings.Join(files, "\t"))
	},
}
