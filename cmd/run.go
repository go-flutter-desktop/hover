package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// TODO: make configurable in case port is taken
const defaultObservatoryPort = "50300"

func init() {
	runCmd.Flags().StringVarP(&buildTargetMainDart, "target", "t", "lib/main.dart", "The main entry-point file of the application.")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Build and start a desktop release, with hot-reload support",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := assertInFlutterProject()

		// Can only run on host OS
		targetOS := runtime.GOOS

		build(projectName, targetOS, []string{"--observatory-port=50300"})
		runAndAttach(projectName, targetOS)
	},
}

func runAndAttach(projectName string, targetOS string) {
	cmdApp := exec.Command(dotSlash + filepath.Join("desktop", "build", "outputs", targetOS, projectName))
	cmdApp.Stderr = os.Stderr
	cmdApp.Stdout = os.Stdout
	err := cmdApp.Start()
	if err != nil {
		fmt.Printf("failed to start app '%s': %v\n", projectName, err)
		os.Exit(1)
	}
	go func() {
		err = cmdApp.Wait()
		if err != nil {
			fmt.Printf("app '%s' exited with error: %v\n", projectName, err)
			os.Exit(cmdApp.ProcessState.ExitCode())
		}
		fmt.Printf("app '%s' exited.\n", projectName)
		os.Exit(0)
	}()

	cmdFlutterAttach := exec.Command("flutter", "attach",
		"--target", buildTargetMainDart,
		"--debug-port", "50300",
		"--device-id", "flutter-tester",
	)
	cmdFlutterAttach.Stdin = os.Stdin
	cmdFlutterAttach.Stdout = os.Stdout
	cmdFlutterAttach.Stderr = os.Stderr
	err = cmdFlutterAttach.Run()
	if err != nil {
		fmt.Printf("flutter attach failed: %v\n", err)
		os.Exit(1)
	}
}
