package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/cmd/packaging"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/config"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

var (
	runObservatoryPort string
	runInitialRoute    string
)

func init() {
	initCompileFlags(runCmd)

	runCmd.Flags().StringVar(&runInitialRoute, "route", "", "Which route to load when running the app.")
	runCmd.Flags().StringVarP(&runObservatoryPort, "observatory-port", "", "50300", "The observatory port used to connect hover to VM services (hot-reload/debug/..)")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Build and start a desktop release, with hot-reload support",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := pubspec.GetPubSpec().Name
		assertHoverInitialized()

		// Can only run on host OS
		targetOS := runtime.GOOS

		initBuildParameters(targetOS, build.DebugMode)
		subcommandBuild(targetOS, packaging.NoopTask, []string{
			"--observatory-port=" + runObservatoryPort,
			"--enable-service-port-fallback",
			"--disable-service-auth-codes",
		})

		log.Infof("Build finished, starting app...")
		runAndAttach(projectName, targetOS)
	},
}

func runAndAttach(projectName string, targetOS string) {
	cmdApp := exec.Command(build.OutputBinaryPath(config.GetConfig().GetExecutableName(projectName), targetOS, buildOrRunMode))
	cmdApp.Env = append(os.Environ(),
		"GOFLUTTER_ROUTE="+runInitialRoute)
	cmdFlutterAttach := exec.Command("flutter", "attach")

	stdoutApp, err := cmdApp.StdoutPipe()
	if err != nil {
		log.Errorf("Unable to create stdout pipe on app: %v", err)
		os.Exit(1)
	}
	stderrApp, err := cmdApp.StderrPipe()
	if err != nil {
		log.Errorf("Unable to create stderr pipe on app: %v", err)
		os.Exit(1)
	}

	regexObservatory := regexp.MustCompile(`Observatory\slistening\son\s(http:[^:]*:\d*/)`)

	// asynchronously read the stdout to catch the debug-uri
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			text := scanner.Text()
			fmt.Println(text)
			match := regexObservatory.FindStringSubmatch(text)
			if len(match) == 2 {
				log.Infof("Connecting hover to '%s' for hot reload", projectName)
				startHotReloadProcess(cmdFlutterAttach, buildOrRunFlutterTarget, match[1])
				break
			}
		}
		// echo command Stdout to terminal
		io.Copy(os.Stdout, stdoutApp)
	}(stdoutApp)

	// Non-blockingly echo command stderr to terminal
	go io.Copy(os.Stderr, stderrApp)

	log.Infof("Running %s in %s mode", projectName, buildOrRunMode.Name)
	err = cmdApp.Start()
	if err != nil {
		log.Errorf("Failed to start app '%s': %v", projectName, err)
		os.Exit(1)
	}

	err = cmdApp.Wait()
	if err != nil {
		log.Errorf("App '%s' exited with error: %v", projectName, err)
		os.Exit(cmdApp.ProcessState.ExitCode())
	}
	log.Infof("App '%s' exited.", projectName)
	log.Printf("Closing the flutter attach sub process..")
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
		log.Warnf("The command 'flutter attach' failed: %v hot reload disabled", err)
	}
}
