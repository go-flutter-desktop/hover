package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/spf13/cobra"
)

// TODO: make configurable in case port is taken
const defaultObservatoryPort = "50300"

func init() {
	runCmd.Flags().StringVarP(&buildTargetMainDart, "target", "t", "lib/main_desktop.dart", "The main entry-point file of the application.")
	runCmd.Flags().StringVarP(&buildTargetManifest, "manifest", "m", "pubspec.yaml", "Flutter manifest file of the application.")
	runCmd.Flags().StringVarP(&buildTargetBranch, "branch", "b", "@master", "The go-flutter-desktop/go-flutter branch to use when building the embedder")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Build and start a desktop release, with hot-reload support",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := assertInFlutterProject()
		assertHoverInitialized()

		// Can only run on host OS
		targetOS := runtime.GOOS

		build(projectName, targetOS, []string{"--observatory-port=50300"})
		runAndAttach(projectName, targetOS)
	},
}

func runAndAttach(projectName string, targetOS string) {
	cmdApp := exec.Command(dotSlash + filepath.Join("desktop", "build", "outputs", targetOS, projectName))
	cmdApp.Stderr = os.Stderr

	// Create stdout, streams to parse the debug-uri of the flutter app.
	// debug-uri is used for hotreloading.
	stdoutApp, err := cmdApp.StdoutPipe()
	if err == nil {
		re := regexp.MustCompile("(?:http:\\/\\/)[^:]*:50300\\/[^\\/]*\\/")

		go func(reader io.Reader) {
			scanner := bufio.NewScanner(reader)
			debugUriFound := false
			for scanner.Scan() {
				fmt.Println(scanner.Text()) // defualt stdout
				if !debugUriFound {
					match := re.FindStringSubmatch(scanner.Text())
					if len(match) == 1 {
						debugUriFound = true
						startHotReloadProcess(buildTargetMainDart, match[0])
					}
				}
			}
		}(stdoutApp)
	} else {
		fmt.Printf("unable to parse flutter debuger'%s' failed with error: %v Hotreload disabled\n", projectName, err)
		cmdApp.Stdout = os.Stdout
	}

	err = cmdApp.Start()
	if err != nil {
		fmt.Printf("failed to start app '%s': %v\n", projectName, err)
		os.Exit(1)
	}

	err = cmdApp.Wait()
	if err != nil {
		fmt.Printf("app '%s' exited with error: %v\n", projectName, err)
		os.Exit(cmdApp.ProcessState.ExitCode())
	}
	fmt.Printf("app '%s' exited.\n", projectName)
	os.Exit(0)
}

func startHotReloadProcess(buildTargetMainDart string, uri string) {
	cmdFlutterAttach := exec.Command("flutter", "attach")
	cmdFlutterAttach.Stdin = os.Stdin
	cmdFlutterAttach.Stdout = os.Stdout
	cmdFlutterAttach.Stderr = os.Stderr

	cmdFlutterAttach.Args = []string{
		"flutter", "attach",
		"--target", buildTargetMainDart,
		"--device-id", "flutter-tester",
		"--debug-uri", uri,
	}
	err := cmdFlutterAttach.Run()
	if err != nil {
		fmt.Printf("flutter attach failed: %v Hotreload disabled\n", err)
	} else {
		defer cmdFlutterAttach.Process.Kill()
	}
}
