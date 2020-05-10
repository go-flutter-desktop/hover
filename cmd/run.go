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

	"github.com/go-flutter-desktop/hover/cmd/packaging"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/config"
	"github.com/go-flutter-desktop/hover/internal/enginecache"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

var (
	runObservatoryPort   string
	runInitialRoute      string
	runOmitEmbedder      bool
	runOmitFlutterBundle bool
	runDocker            bool
)

func init() {
	runCmd.Flags().StringVarP(&buildTarget, "target", "t", config.BuildTargetDefault, "The main entry-point file of the application.")
	runCmd.Flags().StringVarP(&buildGoFlutterBranch, "branch", "b", config.BuildBranchDefault, "The 'go-flutter' version to use. (@master or @v0.20.0 for example)")
	// TODO: The variable buildCachePath is set twice, once during the setup of
	// buildCmd, and once during the setup of runCmd. The last of the two will
	// override the value with it's default, which leads to strange problems.
	runCmd.Flags().StringVar(&buildCachePath, "cache-path", enginecache.DefaultCachePath(), "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll (defaults to the standard user cache directory)")
	runCmd.PersistentFlags().StringVar(&buildEngineVersion, "engine-version", "", "The flutter engine version to use.")
	runCmd.Flags().StringVar(&buildOpenGlVersion, "opengl", config.BuildOpenGlVersionDefault, "The OpenGL version specified here is only relevant for external texture plugin (i.e. video_plugin).\nIf 'none' is provided, texture won't be supported. Note: the Flutter Engine still needs a OpenGL compatible context.")

	runCmd.Flags().StringVar(&runInitialRoute, "route", "", "Which route to load when running the app.")
	runCmd.Flags().StringVarP(&runObservatoryPort, "observatory-port", "", "50300", "The observatory port used to connect hover to VM services (hot-reload/debug/..)")
	runCmd.Flags().BoolVar(&runOmitFlutterBundle, "omit-flutter", false, "Don't (re)compile the current Flutter project, useful when only working with Golang code (plugin)")
	runCmd.Flags().BoolVar(&runOmitEmbedder, "omit-embedder", false, "Don't (re)compile 'go-flutter' source code, useful when only working with Dart code")
	runCmd.Flags().BoolVar(&runDocker, "docker", false, "Execute the go build in a docker container. The Flutter build is always run locally")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Build and start a desktop release, with hot-reload support",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := pubspec.GetPubSpec().Name
		assertHoverInitialized()

		// ensure we have something to build
		if runOmitEmbedder && runOmitFlutterBundle {
			log.Errorf("Flags omit-embedder and omit-flutter are not compatible.")
			os.Exit(1)
		}

		// Can only run on host OS
		targetOS := runtime.GOOS

		// forcefully enable --debug as it is not optional for 'hover run'
		buildDebug = true

		if runOmitFlutterBundle {
			log.Infof("Omiting flutter build bundle")
		} else {
			// TODO: cleaning can't be enabled because it would break when users --omit-embedder.
			// cleanBuildOutputsDir(targetOS)
			buildFlutterBundle(targetOS)
		}
		if runOmitEmbedder {
			log.Infof("Omiting build the embedder")
		} else {
			vmArguments := []string{"--observatory-port=" + runObservatoryPort, "--enable-service-port-fallback", "--disable-service-auth-codes"}
			if runDocker {
				var buildFlags []string
				buildFlags = append(buildFlags, commonFlags()...)
				buildFlags = append(buildFlags, []string{
					"--skip-flutter-build-bundle",
					"--skip-engine-download",
					"--debug",
				}...)
				dockerHoverBuild(targetOS, packaging.NoopTask, buildFlags, vmArguments)
			} else {
				buildGoBinary(targetOS, vmArguments)
			}
		}
		log.Infof("Build finished, starting app...")
		runAndAttach(projectName, targetOS)
	},
}

func runAndAttach(projectName string, targetOS string) {
	cmdApp := exec.Command(dotSlash + filepath.Join(build.BuildPath, "build", "outputs", targetOS, config.GetConfig().GetExecutableName(projectName)))
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
				startHotReloadProcess(cmdFlutterAttach, buildTarget, match[1])
				break
			}
		}
		// echo command Stdout to terminal
		io.Copy(os.Stdout, stdoutApp)
	}(stdoutApp)

	// Non-blockingly echo command stderr to terminal
	go io.Copy(os.Stderr, stderrApp)

	log.Infof("Running %s in debug mode", projectName)
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
