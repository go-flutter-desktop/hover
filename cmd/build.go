package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/internal/enginecache"
	"github.com/go-flutter-desktop/hover/internal/versioncheck"
)

var dotSlash = string([]byte{'.', filepath.Separator})

var (
	buildTarget            string
	buildManifest          string
	buildBranch            string
	buildDebug             bool
	buildCachePath         string
	buildOmitEmbedder      bool
	buildOmitFlutterBundle bool
)

const buildPath = "go"

const mingwGccBinName = "x86_64-w64-mingw32-gcc"
const mingwGppBinName = "x86_64-w64-mingw32-g++"
const clangBinName = "o32-clang"

var crossCompile = false
var windowsCrossCompileToolsAvailable = false
var darwinCrossCompileToolsAvailable = false

func init() {
	buildCmd.PersistentFlags().StringVarP(&buildTarget, "target", "t", "lib/main_desktop.dart", "The main entry-point file of the application.")
	buildCmd.PersistentFlags().StringVarP(&buildManifest, "manifest", "m", "pubspec.yaml", "Flutter manifest file of the application.")
	buildCmd.PersistentFlags().StringVarP(&buildBranch, "branch", "b", "", "The 'go-flutter' version to use. (@master or @v0.20.0 for example)")
	buildCmd.PersistentFlags().BoolVar(&buildDebug, "debug", false, "Build a debug version of the app.")
	buildCmd.PersistentFlags().StringVarP(&buildCachePath, "cache-path", "", "", "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll (defaults to the standard user cache directory)")
	buildCmd.AddCommand(buildLinuxCmd)
	buildCmd.AddCommand(buildLinuxSnapCmd)
	buildCmd.AddCommand(buildLinuxDebCmd)
	buildCmd.AddCommand(buildDarwinCmd)
	buildCmd.AddCommand(buildWindowsCmd)
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a desktop release",
}

var buildLinuxCmd = &cobra.Command{
	Use:   "linux",
	Short: "Build a desktop release for linux",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := getPubSpec().Name
		assertHoverInitialized()

		build(projectName, "linux", nil)
	},
}

var buildLinuxSnapCmd = &cobra.Command{
	Use:   "linux-snap",
	Short: "Build a desktop release for linux and package it for snap",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := getPubSpec().Name
		assertHoverInitialized()
		assertPackagingFormatInitialized("linux-snap")

		build(projectName, "linux", nil)
		buildLinuxSnap(projectName)
	},
}

var buildLinuxDebCmd = &cobra.Command{
	Use:   "linux-deb",
	Short: "Build a desktop release for linux and package it for deb",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := getPubSpec().Name
		assertHoverInitialized()
		assertPackagingFormatInitialized("linux-deb")

		build(projectName, "linux", nil)
		buildLinuxDeb(projectName)
	},
}

var buildDarwinCmd = &cobra.Command{
	Use:   "darwin",
	Short: "Build a desktop release for darwin",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := getPubSpec().Name
		assertHoverInitialized()

		build(projectName, "darwin", nil)
	},
}

var buildWindowsCmd = &cobra.Command{
	Use:   "windows",
	Short: "Build a desktop release for windows",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := getPubSpec().Name
		assertHoverInitialized()

		build(projectName, "windows", nil)
	},
}

func outputDirectoryPath(targetOS string) string {
	outputDirectoryPath, err := filepath.Abs(filepath.Join(buildPath, "build", "outputs", targetOS))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for output directory: %v\n", err)
		os.Exit(1)
	}
	if _, err := os.Stat(outputDirectoryPath); os.IsNotExist(err) {
		err = os.MkdirAll(outputDirectoryPath, 0775)
		if err != nil {
			fmt.Printf("hover: Failed to create output directory %s: %v\n", outputDirectoryPath, err)
			os.Exit(1)
		}
	}
	return outputDirectoryPath
}

func outputBinaryName(projectName string, targetOS string) string {
	var outputBinaryName = projectName
	switch targetOS {
	case "darwin":
		// no special filename
	case "linux":
		// no special filename
	case "windows":
		outputBinaryName += ".exe"
	default:
		fmt.Printf("hover: Target platform %s is not supported.\n", targetOS)
		os.Exit(1)
	}
	return outputBinaryName
}

func outputBinaryPath(projectName string, targetOS string) string {
	outputBinaryPath := filepath.Join(outputDirectoryPath(targetOS), outputBinaryName(projectName, targetOS))
	return outputBinaryPath
}

func dockerBuild(projectName string, targetOS string, vmArguments []string) {
	dockerBin, err := exec.LookPath("docker")
	if err != nil {
		fmt.Printf("hover: Failed to lookup `docker` executable. Please install Docker.\nhttps://docs.docker.com/install/")
		os.Exit(1)
	}
	tmpDir, err := ioutil.TempDir("", "hover-build-cc")
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		fmt.Printf("Cannot get the path for the system cache directory %s", err)
		os.Exit(1)
	}
	goPath := filepath.Join(userCacheDir, "hover-cc")
	err = os.MkdirAll(goPath, 0755)
	if err != nil {
		fmt.Printf("Cannot create the hover-cc GOPATH under the system cache directory %s", err)
		os.Exit(1)
	}
	pkgRootDir, err := filepath.Abs(buildPath)
	if err != nil {
		fmt.Printf("Cannot get the path for current directory %s", err)
		os.Exit(1)
	}
	var engineCachePath string
	if buildCachePath != "" {
		engineCachePath = enginecache.ValidateOrUpdateEngineAtPath(targetOS, buildCachePath)
	} else {
		engineCachePath = enginecache.ValidateOrUpdateEngine(targetOS)
	}
	dockerFilePath, err := filepath.Abs(filepath.Join(tmpDir, "Dockerfile"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for Dockerfile %s: %v\n", dockerFilePath, err)
		os.Exit(1)
	}

	dockerFile, err := os.Create(dockerFilePath)
	if err != nil {
		fmt.Printf("hover: Failed to create Dockerfile %s: %v\n", dockerFilePath, err)
		os.Exit(1)
	}
	dockerFileContent := []string{
		"FROM dockercore/golang-cross",
		"WORKDIR /app",
		"CMD " + strings.Join(buildCommand(targetOS, vmArguments, "build/outputs/"+targetOS+"/"+outputBinaryName(projectName, targetOS)), " "),
	}
	for _, env := range buildEnv(targetOS, "/engine") {
		dockerFileContent = append(dockerFileContent, "ENV "+env)
	}

	for _, line := range dockerFileContent {
		if _, err := dockerFile.WriteString(line + "\n"); err != nil {
			fmt.Printf("hover: Could not write Dockerfile: %v\n", err)
			os.Exit(1)
		}
	}
	err = dockerFile.Close()
	if err != nil {
		fmt.Printf("hover: Could not close Dockerfile: %v\n", err)
		os.Exit(1)
	}
	dockerBuildCmd := exec.Command(dockerBin, "build", "-t", "hover-build-cc", ".")
	dockerBuildCmd.Stderr = os.Stderr
	dockerBuildCmd.Dir = tmpDir
	err = dockerBuildCmd.Run()
	if err != nil {
		fmt.Printf("hover: Docker build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("hover: Cross-Compiling 'go-flutter' and plugins using docker")

	outputPath, err := filepath.Abs(filepath.Join(buildPath, "build", "outputs"))
	dockerRunCmd := exec.Command(dockerBin, "run", "-e", "USERID=$UID", "-v", goPath+":/go", "-v", pkgRootDir+":/app", "-v", engineCachePath+":/engine", "-v", outputPath+":/app/build/outputs", "-v", filepath.Join(userCacheDir, "go-build")+":/cache", "hover-build-cc")
	dockerRunCmd.Stderr = os.Stderr
	dockerRunCmd.Stdout = os.Stdout
	dockerRunCmd.Dir = tmpDir
	err = dockerRunCmd.Run()
	if err != nil {
		fmt.Printf("hover: Docker run failed: %v\n", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpDir)
	if err != nil {
		fmt.Printf("hover: Could not remove temporary build folder: %v\n", err)
		os.Exit(1)
	}
}

func build(projectName string, targetOS string, vmArguments []string) {
	mingwGccBin, err := exec.LookPath(mingwGccBinName)
	mingwGppBin, err := exec.LookPath(mingwGppBinName)
	clangBin, err := exec.LookPath(clangBinName)
	crossCompile = targetOS != runtime.GOOS
	if crossCompile {
		fmt.Printf("hover: Cross-compiling for %s using docker is very experimental\n", targetOS)
		if mingwGccBin != "" && mingwGppBin != "" {
			windowsCrossCompileToolsAvailable = true
		}
		if clangBin != "" {
			darwinCrossCompileToolsAvailable = true
		}
	}
	var engineCachePath string
	if buildCachePath != "" {
		engineCachePath = enginecache.ValidateOrUpdateEngineAtPath(targetOS, buildCachePath)
	} else {
		engineCachePath = enginecache.ValidateOrUpdateEngine(targetOS)
	}

	if !buildOmitFlutterBundle && !buildOmitEmbedder {
		err := os.RemoveAll(outputDirectoryPath(targetOS))
		fmt.Printf("hover: Cleaning the build directory\n")
		if err != nil {
			fmt.Printf("hover: failed to clean output directory %s: %v\n", outputDirectoryPath(targetOS), err)
			os.Exit(1)
		}
	}

	err = os.MkdirAll(outputDirectoryPath(targetOS), 0775)
	if err != nil {
		fmt.Printf("hover: failed to create output directory %s: %v\n", outputDirectoryPath(targetOS), err)
		os.Exit(1)
	}

	cmdCheckFlutter := exec.Command(flutterBin, "--version")
	cmdCheckFlutterOut, err := cmdCheckFlutter.Output()
	if err != nil {
		fmt.Printf("hover: failed to check your flutter channel: %v\n", err)
	} else {
		re := regexp.MustCompile("•\\schannel\\s(\\w*)\\s•")

		match := re.FindStringSubmatch(string(cmdCheckFlutterOut))
		if len(match) >= 2 {
			ignoreWarning := os.Getenv("HOVER_IGNORE_CHANNEL_WARNING")
			if match[1] != "beta" && ignoreWarning != "true" {
				fmt.Println("hover: ⚠ The go-flutter project tries to stay compatible with the beta channel of Flutter.")
				fmt.Println("hover: ⚠     It's advised to use the beta channel. ($ flutter channel beta)")
			}
		} else {
			fmt.Printf("hover: failed to check your flutter channel: Unrecognized output format")
		}
	}

	var trackWidgetCreation string
	if buildDebug {
		trackWidgetCreation = "--track-widget-creation"
	}

	cmdFlutterBuild := exec.Command(flutterBin, "build", "bundle",
		"--asset-dir", filepath.Join(outputDirectoryPath(targetOS), "flutter_assets"),
		"--target", buildTarget,
		"--manifest", buildManifest,
		trackWidgetCreation,
	)
	cmdFlutterBuild.Stderr = os.Stderr
	cmdFlutterBuild.Stdout = os.Stdout

	if !buildOmitFlutterBundle {
		fmt.Printf("hover: Bundling flutter app\n")
		err = cmdFlutterBuild.Run()
		if err != nil {
			fmt.Printf("hover: Flutter build failed: %v\n", err)
			os.Exit(1)
		}
	}

	var engineFile string
	switch targetOS {
	case "darwin":
		engineFile = "FlutterEmbedder.framework"
	case "linux":
		engineFile = "libflutter_engine.so"
	case "windows":
		engineFile = "flutter_engine.dll"
	}

	outputEngineFile := filepath.Join(outputDirectoryPath(targetOS), engineFile)
	err = copy.Copy(
		filepath.Join(engineCachePath, engineFile),
		outputEngineFile,
	)
	if err != nil {
		fmt.Printf("hover: Failed to copy %s: %v\n", engineFile, err)
		os.Exit(1)
	}
	if !buildDebug && targetOS == "linux" {
		err = exec.Command("strip", "-s", outputEngineFile).Run()
		if err != nil {
			fmt.Printf("Failed to strip %s: %v\n", outputEngineFile, err)
			os.Exit(1)
		}
	}

	err = copy.Copy(
		filepath.Join(engineCachePath, "artifacts", "icudtl.dat"),
		filepath.Join(outputDirectoryPath(targetOS), "icudtl.dat"),
	)
	if err != nil {
		fmt.Printf("hover: Failed to copy icudtl.dat: %v\n", err)
		os.Exit(1)
	}

	err = copy.Copy(
		filepath.Join(buildPath, "assets"),
		filepath.Join(outputDirectoryPath(targetOS), "assets"),
	)
	if err != nil {
		fmt.Printf("hover: Failed to copy %s/assets: %v\n", buildPath, err)
		os.Exit(1)
	}

	if buildOmitEmbedder {
		// Omit the 'go-flutter' build
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("hover: Failed to get working dir: %v\n", err)
		os.Exit(1)
	}

	if buildBranch == "" {

		currentTag, err := versioncheck.CurrentGoFlutterTag(filepath.Join(wd, buildPath))
		if err != nil {
			fmt.Printf("hover: %v\n", err)
			os.Exit(1)
		}

		semver, err := version.NewSemver(currentTag)
		if err != nil {
			fmt.Printf("hover: faild to parse 'go-flutter' semver: %v\n", err)
			os.Exit(1)
		}

		if semver.Prerelease() != "" {
			fmt.Printf("hover: Upgrade 'go-flutter' to the latest release\n")
			// no buildBranch provided and currentTag isn't a release,
			// force update. (same behaviour as previous version of hover).
			err = upgradeGoFlutter(targetOS, engineCachePath)
			if err != nil {
				// the upgrade can fail silently
				fmt.Printf("hover: Upgrade ignored, current 'go-flutter' version: %s\n", currentTag)
			}
		} else {
			// when the buildBranch is empty and the currentTag is a release.
			// Check if the 'go-flutter' needs updates.
			versioncheck.CheckFoGoFlutterUpdate(filepath.Join(wd, buildPath), currentTag)
		}

	} else {
		fmt.Printf("hover: Downloading 'go-flutter' %s\n", buildBranch)

		// when the buildBranch is set, fetch the go-flutter branch version.
		err = upgradeGoFlutter(targetOS, engineCachePath)
		if err != nil {
			os.Exit(1)
		}
	}

	if crossCompile {
		dockerBuild(projectName, targetOS, vmArguments)
		return
	}

	buildCommandString := buildCommand(targetOS, vmArguments, outputBinaryPath(projectName, targetOS))
	cmdGoBuild := exec.Command(buildCommandString[0], buildCommandString[1:]...)
	cmdGoBuild.Dir = filepath.Join(wd, buildPath)
	cmdGoBuild.Env = append(os.Environ(),
		buildEnv(targetOS, engineCachePath)...,
	)

	cmdGoBuild.Stderr = os.Stderr
	cmdGoBuild.Stdout = os.Stdout

	fmt.Printf("hover: Compiling 'go-flutter' and plugins\n")
	err = cmdGoBuild.Run()
	if err != nil {
		fmt.Printf("hover: Go build failed: %v\n", err)
		os.Exit(1)
	}
}

func buildEnv(targetOS string, engineCachePath string) []string {
	var cgoLdflags string
	switch targetOS {
	case "darwin":
		cgoLdflags = fmt.Sprintf("-F%s -Wl,-rpath,@executable_path", engineCachePath)
	case "linux":
		cgoLdflags = fmt.Sprintf("-L%s", engineCachePath)
	case "windows":
		cgoLdflags = fmt.Sprintf("-L%s", engineCachePath)
	default:
		fmt.Printf("hover: Target platform %s is not supported, cgo_ldflags not implemented.\n", targetOS)
		os.Exit(1)
	}
	cgoLdflagsString := ""
	if crossCompile {
		cgoLdflagsString = "\""
	}
	cgoLdflagsString = cgoLdflagsString + cgoLdflags
	if crossCompile {
		cgoLdflagsString = cgoLdflagsString + "\""
	}
	env := []string{
		"GO111MODULE=on",
		"CGO_LDFLAGS=" + cgoLdflagsString,
		"GOOS=" + targetOS,
		"GOARCH=amd64",
		"CGO_ENABLED=1",
	}
	if crossCompile {
		env = append(env,
			"GOCACHE=/cache",
		)
		if targetOS == "windows" {
			env = append(env,
				"CC="+mingwGccBinName,
				"CXX="+mingwGppBinName,
			)
		}
		if targetOS == "darwin" {
			env = append(env,
				"CC="+clangBinName,
				"CXX="+clangBinName,
			)
		}
	}
	return env
}

func buildCommand(targetOS string, vmArguments []string, outputBinaryPath string) []string {
	var ldflags []string
	if !buildDebug {
		vmArguments = append(vmArguments, "--disable-dart-asserts")
		vmArguments = append(vmArguments, "--disable-observatory")

		if targetOS == "windows" {
			ldflags = append(ldflags, "-H=windowsgui")
		}
		ldflags = append(ldflags, "-s")
		ldflags = append(ldflags, "-w")
	}
	ldflags = append(ldflags, fmt.Sprintf("-X main.vmArguments=%s", strings.Join(vmArguments, ";")))
	outputCommand := []string{
		"go",
		"build",
		"-o", outputBinaryPath,
		"-v",
	}
	if crossCompile {
		outputCommand = append(outputCommand, fmt.Sprintf("-ldflags=\"%s\"", strings.Join(ldflags, " ")))
	} else {
		outputCommand = append(outputCommand, fmt.Sprintf("-ldflags=%s", strings.Join(ldflags, " ")))
	}
	outputCommand = append(outputCommand, dotSlash+"cmd")
	return outputCommand
}
