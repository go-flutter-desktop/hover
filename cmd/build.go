package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/internal/enginecache"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/versioncheck"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
	"github.com/go-flutter-desktop/hover/cmd/packaging"
)

var dotSlash = string([]byte{'.', filepath.Separator})

var (
	buildTarget            string
	buildBranch            string
	buildDebug             bool
	buildCachePath         string
	buildOmitEmbedder      bool
	buildOmitFlutterBundle bool
	buildOpenGlVersion     string
	buildDocker            bool
)

const mingwGccBinName = "x86_64-w64-mingw32-gcc"
const clangBinName = "o32-clang"

var crossCompile = false
var engineCachePath string

func init() {
	buildCmd.PersistentFlags().StringVarP(&buildTarget, "target", "t", "lib/main_desktop.dart", "The main entry-point file of the application.")
	buildCmd.PersistentFlags().StringVarP(&buildBranch, "branch", "b", "", "The 'go-flutter' version to use. (@master or @v0.20.0 for example)")
	buildCmd.PersistentFlags().BoolVar(&buildDebug, "debug", false, "Build a debug version of the app.")
	buildCmd.PersistentFlags().StringVarP(&buildCachePath, "cache-path", "", "", "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll (defaults to the standard user cache directory)")
	buildCmd.PersistentFlags().StringVar(&buildOpenGlVersion, "opengl", "3.3", "The OpenGL version specified here is only relevant for external texture plugin (i.e. video_plugin).\nIf 'none' is provided, texture won't be supported. Note: the Flutter Engine still needs a OpenGL compatible context.")
	buildCmd.PersistentFlags().BoolVar(&buildDocker, "docker", false, "Compile in Docker container only. No need to install go")
	buildCmd.AddCommand(buildLinuxCmd)
	buildCmd.AddCommand(buildLinuxSnapCmd)
	buildCmd.AddCommand(buildLinuxDebCmd)
	buildCmd.AddCommand(buildDarwinCmd)
	buildCmd.AddCommand(buildDarwinBundleCmd)
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
		assertHoverInitialized()

		buildNormal("linux", nil)
	},
}

var buildLinuxSnapCmd = &cobra.Command{
	Use:   "linux-snap",
	Short: "Build a desktop release for linux and package it for snap",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.AssertPackagingFormatInitialized("linux-snap")

		if !packaging.DockerInstalled() {
			os.Exit(1)
		}

		buildNormal("linux", nil)
		packaging.BuildLinuxSnap()
	},
}

var buildLinuxDebCmd = &cobra.Command{
	Use:   "linux-deb",
	Short: "Build a desktop release for linux and package it for deb",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.AssertPackagingFormatInitialized("linux-deb")

		if !packaging.DockerInstalled() {
			os.Exit(1)
		}

		buildNormal("linux", nil)
		packaging.BuildLinuxDeb()
	},
}

var buildDarwinCmd = &cobra.Command{
	Use:   "darwin",
	Short: "Build a desktop release for darwin",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()

		buildNormal("darwin", nil)
	},
}

var buildDarwinBundleCmd = &cobra.Command{
	Use:   "darwin-bundle",
	Short: "Build a desktop release for darwin and package it for OSX bundle",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.AssertPackagingFormatInitialized("darwin-bundle")

		if !packaging.DockerInstalled() {
			os.Exit(1)
		}

		buildNormal("darwin", nil)
		packaging.BuildDarwinBundle()
	},
}

var buildWindowsCmd = &cobra.Command{
	Use:   "windows",
	Short: "Build a desktop release for windows",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()

		buildNormal("windows", nil)
	},
}

func buildInDocker(targetOS string, vmArguments []string) {
	crossCompilingDir, err := filepath.Abs(filepath.Join(build.BuildPath, "cross-compiling"))
	err = os.MkdirAll(crossCompilingDir, 0755)
	if err != nil {
		log.Errorf("Cannot create the cross-compiling directory: %v", err)
		os.Exit(1)
	}
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Errorf("Cannot get the path for the system cache directory: %v", err)
		os.Exit(1)
	}
	goPath := filepath.Join(userCacheDir, "hover-cc")
	err = os.MkdirAll(goPath, 0755)
	if err != nil {
		log.Errorf("Cannot create the hover-cc GOPATH under the system cache directory: %v", err)
		os.Exit(1)
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Errorf("Cannot get the path for current directory %s", err)
		os.Exit(1)
	}
	dockerFilePath, err := filepath.Abs(filepath.Join(crossCompilingDir, "Dockerfile"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for Dockerfile %s: %v", dockerFilePath, err)
		os.Exit(1)
	}
	if _, err := os.Stat(dockerFilePath); os.IsNotExist(err) {
		dockerFile, err := os.Create(dockerFilePath)
		if err != nil {
			log.Errorf("Failed to create Dockerfile %s: %v", dockerFilePath, err)
			os.Exit(1)
		}
		dockerFileContent := []string{
			"FROM dockercore/golang-cross",
			"RUN apt-get update && apt-get install libgl1-mesa-dev xorg-dev -y",
		}

		for _, line := range dockerFileContent {
			if _, err := dockerFile.WriteString(line + "\n"); err != nil {
				log.Errorf("Could not write Dockerfile: %v", err)
				os.Exit(1)
			}
		}
		err = dockerFile.Close()
		if err != nil {
			log.Errorf("Could not close Dockerfile: %v", err)
			os.Exit(1)
		}
		log.Infof("A Dockerfile for cross-compiling for %s has been created at %s. You can add it to git.", targetOS, filepath.Join(build.BuildPath, "cross-compiling", targetOS))
	}
	dockerBuildCmd := exec.Command(build.DockerBin, "build", "-t", "hover-build-cc", ".")
	dockerBuildCmd.Stdout = os.Stdout
	dockerBuildCmd.Stderr = os.Stderr
	dockerBuildCmd.Dir = crossCompilingDir
	err = dockerBuildCmd.Run()
	if err != nil {
		log.Errorf("Docker build failed: %v", err)
		os.Exit(1)
	}

	log.Infof("Cross-Compiling 'go-flutter' and plugins using docker")

	u, err := user.Current()
	if err != nil {
		log.Errorf("Couldn't get current user: %v", err)
		os.Exit(1)
	}
	args := []string{
		"run",
		"-w", "/app/go",
		"-v", goPath + ":/go",
		"-v", wd + ":/app",
		"-v", engineCachePath + ":/engine",
		"-v", filepath.Join(userCacheDir, "go-build") + ":/cache",
	}
	for _, env := range buildEnv(targetOS, "/engine") {
		args = append(args, "-e", env)
	}
	args = append(args, "hover-build-cc")
	chownStr := ""
	if runtime.GOOS != "windows" {
		chownStr = fmt.Sprintf(" && chown %s:%s build/ -R", u.Uid, u.Gid)
	}
	args = append(args, "bash", "-c", fmt.Sprintf("%s%s", strings.Join(buildCommand(targetOS, vmArguments, "build/outputs/"+targetOS+"/"+build.OutputBinaryName(pubspec.GetPubSpec().Name, targetOS)), " "), chownStr))
	dockerRunCmd := exec.Command(build.DockerBin, args...)
	dockerRunCmd.Stderr = os.Stderr
	dockerRunCmd.Stdout = os.Stdout
	dockerRunCmd.Dir = crossCompilingDir
	err = dockerRunCmd.Run()
	if err != nil {
		log.Errorf("Docker run failed: %v", err)
		os.Exit(1)
	}
	log.Infof("Successfully cross-compiled for " + targetOS)
}

func buildNormal(targetOS string, vmArguments []string) {
	crossCompile = targetOS != runtime.GOOS
	buildDocker = crossCompile || buildDocker

	if buildCachePath != "" {
		engineCachePath = enginecache.ValidateOrUpdateEngineAtPath(targetOS, buildCachePath)
	} else {
		engineCachePath = enginecache.ValidateOrUpdateEngine(targetOS)
	}

	if !buildOmitFlutterBundle && !buildOmitEmbedder {
		err := os.RemoveAll(build.OutputDirectoryPath(targetOS))
		log.Printf("Cleaning the build directory")
		if err != nil {
			log.Errorf("Failed to clean output directory %s: %v", build.OutputDirectoryPath(targetOS), err)
			os.Exit(1)
		}
	}

	err := os.MkdirAll(build.OutputDirectoryPath(targetOS), 0775)
	if err != nil {
		log.Errorf("Failed to create output directory %s: %v", build.OutputDirectoryPath(targetOS), err)
		os.Exit(1)
	}

	cmdCheckFlutter := exec.Command(build.FlutterBin, "--version")
	cmdCheckFlutterOut, err := cmdCheckFlutter.Output()
	if err != nil {
		log.Warnf("Failed to check your flutter channel: %v", err)
	} else {
		re := regexp.MustCompile("•\\schannel\\s(\\w*)\\s•")

		match := re.FindStringSubmatch(string(cmdCheckFlutterOut))
		if len(match) >= 2 {
			ignoreWarning := os.Getenv("HOVER_IGNORE_CHANNEL_WARNING")
			if match[1] != "beta" && ignoreWarning != "true" {
				log.Warnf("⚠ The go-flutter project tries to stay compatible with the beta channel of Flutter.")
				log.Warnf("⚠     It's advised to use the beta channel: %s", log.Au().Magenta("flutter channel beta"))
			}
		} else {
			log.Warnf("Failed to check your flutter channel: Unrecognized output format")
		}
	}

	var trackWidgetCreation string
	if buildDebug {
		trackWidgetCreation = "--track-widget-creation"
	}

	cmdFlutterBuild := exec.Command(build.FlutterBin, "build", "bundle",
		"--asset-dir", filepath.Join(build.OutputDirectoryPath(targetOS), "flutter_assets"),
		"--target", buildTarget,
		trackWidgetCreation,
	)
	cmdFlutterBuild.Stderr = os.Stderr
	cmdFlutterBuild.Stdout = os.Stdout

	if !buildOmitFlutterBundle {
		log.Infof("Bundling flutter app")
		err = cmdFlutterBuild.Run()
		if err != nil {
			log.Errorf("Flutter build failed: %v", err)
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

	outputEngineFile := filepath.Join(build.OutputDirectoryPath(targetOS), engineFile)
	err = copy.Copy(
		filepath.Join(engineCachePath, engineFile),
		outputEngineFile,
	)
	if err != nil {
		log.Errorf("Failed to copy %s: %v", engineFile, err)
		os.Exit(1)
	}
	if !buildDebug && targetOS == "linux" {
		err = exec.Command("strip", "-s", outputEngineFile).Run()
		if err != nil {
			log.Errorf("Failed to strip %s: %v", outputEngineFile, err)
			os.Exit(1)
		}
	}

	err = copy.Copy(
		filepath.Join(engineCachePath, "artifacts", "icudtl.dat"),
		filepath.Join(build.OutputDirectoryPath(targetOS), "icudtl.dat"),
	)
	if err != nil {
		log.Errorf("Failed to copy icudtl.dat: %v", err)
		os.Exit(1)
	}

	err = copy.Copy(
		filepath.Join(build.BuildPath, "assets"),
		filepath.Join(build.OutputDirectoryPath(targetOS), "assets"),
	)
	if err != nil {
		log.Errorf("Failed to copy %s/assets: %v", build.BuildPath, err)
		os.Exit(1)
	}

	if buildOmitEmbedder {
		// Omit the 'go-flutter' build
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Errorf("Failed to get working dir: %v", err)
		os.Exit(1)
	}

	if buildBranch == "" {
		currentTag, err := versioncheck.CurrentGoFlutterTag(filepath.Join(wd, build.BuildPath))
		if err != nil {
			log.Errorf("%v", err)
			os.Exit(1)
		}

		semver, err := version.NewSemver(currentTag)
		if err != nil {
			log.Errorf("Faild to parse 'go-flutter' semver: %v", err)
			os.Exit(1)
		}

		if semver.Prerelease() != "" {
			log.Infof("Upgrading 'go-flutter' to the latest release")
			// no buildBranch provided and currentTag isn't a release,
			// force update. (same behaviour as previous version of hover).
			err = upgradeGoFlutter(targetOS, engineCachePath)
			if err != nil {
				// the upgrade can fail silently
				log.Warnf("Upgrade ignored, current 'go-flutter' version: %s", currentTag)
			}
		} else {
			// when the buildBranch is empty and the currentTag is a release.
			// Check if the 'go-flutter' needs updates.
			versioncheck.CheckForGoFlutterUpdate(filepath.Join(wd, build.BuildPath), currentTag)
		}

	} else {
		log.Printf("Downloading 'go-flutter' %s", buildBranch)

		// when the buildBranch is set, fetch the go-flutter branch version.
		err = upgradeGoFlutter(targetOS, engineCachePath)
		if err != nil {
			os.Exit(1)
		}
	}

	if buildOpenGlVersion == "none" {
		log.Warnf("The '--opengl=none' flag makes go-flutter incompatible with texture plugins!")
	}

	if buildDocker {
		if crossCompile {
			log.Infof("Because %s is not able to compile for %s out of the box, a cross-compiling container is used", runtime.GOOS, targetOS)
		}
		buildInDocker(targetOS, vmArguments)
		return
	}

	buildCommandString := buildCommand(targetOS, vmArguments, build.OutputBinaryPath(pubspec.GetPubSpec().Name, targetOS))
	cmdGoBuild := exec.Command(buildCommandString[0], buildCommandString[1:]...)
	cmdGoBuild.Dir = filepath.Join(wd, build.BuildPath)
	cmdGoBuild.Env = append(os.Environ(),
		buildEnv(targetOS, engineCachePath)...,
	)

	cmdGoBuild.Stderr = os.Stderr
	cmdGoBuild.Stdout = os.Stdout

	log.Infof("Compiling 'go-flutter' and plugins")
	err = cmdGoBuild.Run()
	if err != nil {
		log.Errorf("Go build failed: %v", err)
		os.Exit(1)
	}
	log.Infof("Successfully compiled")
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
		log.Errorf("Target platform %s is not supported, cgo_ldflags not implemented.", targetOS)
		os.Exit(1)
	}
	env := []string{
		"GO111MODULE=on",
		"CGO_LDFLAGS=" + cgoLdflags,
		"GOOS=" + targetOS,
		"GOARCH=amd64",
		"CGO_ENABLED=1",
	}
	if buildDocker {
		env = append(env,
			"GOCACHE=/cache",
		)
		if targetOS == "windows" {
			env = append(env,
				"CC="+mingwGccBinName,
			)
		}
		if targetOS == "darwin" {
			env = append(env,
				"CC="+clangBinName,
			)
		}
	}
	return env
}

func buildCommand(targetOS string, vmArguments []string, outputBinaryPath string) []string {
	currentTag, err := versioncheck.CurrentGoFlutterTag(build.BuildPath)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

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
	// overwrite go-flutter build-constants values
	ldflags = append(ldflags, fmt.Sprintf(
		"-X github.com/go-flutter-desktop/go-flutter.ProjectVersion=%s "+
			" -X github.com/go-flutter-desktop/go-flutter.PlatformVersion=%s "+
			" -X github.com/go-flutter-desktop/go-flutter.ProjectName=%s "+
			" -X github.com/go-flutter-desktop/go-flutter.ProjectOrganizationName=%s",
		pubspec.GetPubSpec().Version,
		currentTag,
		pubspec.GetPubSpec().Name,
		androidOrganizationName()))

	outputCommand := []string{
		"go",
		"build",
		"-tags=opengl" + buildOpenGlVersion,
		"-o", outputBinaryPath,
		"-v",
	}
	if buildDocker {
		outputCommand = append(outputCommand, fmt.Sprintf("-ldflags=\"%s\"", strings.Join(ldflags, " ")))
	} else {
		outputCommand = append(outputCommand, fmt.Sprintf("-ldflags=%s", strings.Join(ldflags, " ")))
	}
	outputCommand = append(outputCommand, dotSlash+"cmd")
	return outputCommand
}
