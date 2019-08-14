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
	runCmd.Flags().StringVarP(&buildTarget, "target", "t", "lib/main_desktop.dart", "The main entry-point file of the application.")
	runCmd.Flags().StringVarP(&buildManifest, "manifest", "m", "pubspec.yaml", "Flutter manifest file of the application.")
	buildCmd.Flags().StringVarP(&buildBranch, "branch", "b", "", "The 'go-flutter' version to use. (@master or @v0.20.0 for example)")
	runCmd.Flags().StringVarP(&buildCachePath, "cache-path", "", "", "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll (defaults to the standard user cache directory)")
	runCmd.Flags().BoolVar(&buildOmitEmbedder, "omit-embedder", false, "Don't (re)compile 'go-flutter' source code, useful when only working with Dart code")
	runCmd.Flags().BoolVar(&buildOmitFlutterBundle, "omit-flutter", false, "Don't (re)compile the current Flutter project, useful when only working with Golang code (plugin)")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Build and start a desktop release, with hot-reload support",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := assertInFlutterProject()
		assertHoverInitialized()

		// ensure we have something to build
		if buildOmitEmbedder && buildOmitFlutterBundle {
			fmt.Println("hover: flags omit-embedder and omit-flutter are not compatible.")
			os.Exit(1)
		}

		// Can only run on host OS
		targetOS := runtime.GOOS

		// forcefully enable --debug (which is not an option for `hover run`)
		buildDebug = true

		build(projectName, targetOS, []string{"--observatory-port=50300"})
		runAndAttach(projectName, targetOS)
	},
}

func runAndAttach(projectName string, targetOS string) {
	cmdApp := exec.Command(dotSlash + filepath.Join("desktop", "build", "outputs", targetOS, projectName))
	cmdFlutterAttach := exec.Command("flutter", "attach")

	stdoutApp, err := cmdApp.StdoutPipe()
	if err != nil {
		fmt.Printf("hover: unable to create stdout pipe on app: %v\n", err)
		os.Exit(1)
	}
	stderrApp, err := cmdApp.StderrPipe()
	if err != nil {
		fmt.Printf("hover: unable to create stderr pipe on app: %v\n", err)
		os.Exit(1)
	}

	re := regexp.MustCompile("(?:http:\\/\\/)[^:]*:50300\\/[^\\/]*\\/")

	// Non-blockingly read the stdout to catch the debug-uri
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			text := scanner.Text()
			fmt.Println(text)
			match := re.FindStringSubmatch(text)
			if len(match) == 1 {
				startHotReloadProcess(cmdFlutterAttach, buildTarget, match[0])
				break
			}
		}
		// echo command Stdout to terminal
		io.Copy(os.Stdout, stdoutApp)
	}(stdoutApp)

	// Non-blockingly echo command stderr to terminal
	go io.Copy(os.Stderr, stderrApp)

	fmt.Printf("hover: Running %s in debug mode\n", projectName)
	err = cmdApp.Start()
	if err != nil {
		fmt.Printf("hover: failed to start app '%s': %v\n", projectName, err)
		os.Exit(1)
	}

	err = cmdApp.Wait()
	if err != nil {
		fmt.Printf("hover: app '%s' exited with error: %v\n", projectName, err)
		os.Exit(cmdApp.ProcessState.ExitCode())
	}
	fmt.Printf("hover: app '%s' exited.\n", projectName)
	fmt.Println("hover: closing the flutter attach sub process..")
	cmdFlutterAttach.Wait()
	os.Exit(0)
}

func startHotReloadProcess(cmdFlutterAttach *exec.Cmd, buildTargetMainDart string, uri string) {
	cmdFlutterAttach.Stdin = os.Stdin
	cmdFlutterAttach.Stdout = os.Stdout
	cmdFlutterAttach.Stderr = os.Stderr

	cmdFlutterAttach.Args = []string{
		"flutter", "attach",
		"--target", buildTargetMainDart,
		"--device-id", "flutter-tester",
		"--debug-uri", uri,
	}
	err := cmdFlutterAttach.Start()
	if err != nil {
		fmt.Printf("hover: flutter attach failed: %v Hotreload disabled\n", err)
	}
}
