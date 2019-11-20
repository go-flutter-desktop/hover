package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/cmd/packaging"
	"github.com/go-flutter-desktop/hover/internal/androidmanifest"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/config"
	"github.com/go-flutter-desktop/hover/internal/enginecache"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
	"github.com/go-flutter-desktop/hover/internal/versioncheck"
)

var dotSlash = string([]byte{'.', filepath.Separator})

var (
	buildTargetFile        string
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
	buildCmd.Flags().StringVarP(&buildTargetFile, "target", "t", config.BuildTargetFileDefault, "The main entry-point file of the application.")
	buildCmd.Flags().StringVarP(&buildBranch, "branch", "b", config.BuildBranchDefault, "The `go-flutter` version to use. (@master or @v0.20.0 for example)")
	buildCmd.Flags().BoolVar(&buildDebug, "debug", false, "Build a debug version of the app.")
	buildCmd.Flags().StringVarP(&buildCachePath, "cache-path", "", config.BuildCachePathDefault, "The path that hover uses to cache dependencies such as the Flutter engine .so/.dll (defaults to the standard user cache directory)")
	buildCmd.Flags().StringVar(&buildOpenGlVersion, "opengl", config.BuildOpenGlVersionDefault, "The OpenGL version specified here is only relevant for external texture plugin (i.e. video_plugin).\nIf `none` is provided, texture won`t be supported. Note: the Flutter Engine still needs a OpenGL compatible context.")
	buildCmd.Flags().BoolVar(&buildDocker, "docker", false, "Compile in Docker container only. No need to install go")
	buildCmd.AddCommand(listCmd)
	rootCmd.AddCommand(buildCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all targets",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("All platform targets:")
		for _, buildTarget := range build.PlatformTargets {
			line := fmt.Sprintf("    %s", buildTarget.Platform)
			if buildTarget.Platform != runtime.GOOS {
				if build.DockerBin == "" {
					line += " (docker needs to be installed)"
				}
			}
			log.Infof(line)
		}
		log.Infof("All packaging targets:")
		for _, buildTarget := range build.PackagingTargets {
			line := fmt.Sprintf("    %s-%s", buildTarget.Platform, buildTarget.PackagingFormat)
			_, err := os.Stat(packaging.PackagingFormatPath(buildTarget))
			initialized := err == nil
			if build.DockerBin == "" || !initialized {
				line += " ("
				if build.DockerBin == "" {
					line += "docker needs to be installed"
				}
				if build.DockerBin == "" && !initialized {
					line += "; "
				}
				if !initialized {
					line += "needs to be initialized first"
				}
				line += ")"
			}
			log.Infof(line)
		}
	},
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a desktop release",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a build targets argument")
		}
		return build.AreValidBuildTargets(args[0], false)
	},
	Run: func(cmd *cobra.Command, args []string) {
		buildTargets, err := build.ParseBuildTargets(args[0], false)
		if err != nil {
			log.Errorf("Failed to parse build targets: %v", err)
			os.Exit(1)
		}
		assertHoverInitialized()
		hasLinuxTargets := false
		hasDarwinTargets := false
		hasWindowsTargets := false
		hasCrossCompilingTargets := false
		hasPackagingTargets := false
		for _, buildTarget := range buildTargets {
			if buildTarget.Platform == build.TargetPlatforms.Linux {
				hasLinuxTargets = true
			}
			if buildTarget.Platform == build.TargetPlatforms.Darwin {
				hasDarwinTargets = true
			}
			if buildTarget.Platform == build.TargetPlatforms.Windows {
				hasWindowsTargets = true
			}
			if runtime.GOOS != buildTarget.Platform {
				hasCrossCompilingTargets = true
			}
			if buildTarget.PackagingFormat != "" {
				hasPackagingTargets = true
				packaging.AssertPackagingFormatInitialized(buildTarget)
			}
		}
		if hasCrossCompilingTargets || hasPackagingTargets {
			packaging.AssertDockerInstalled()
		}
		checkForMainDesktop()
		checkFlutter()
		runPluginGet()
		overrideBuildConfig()
		buildFlutterBundle()
		if hasLinuxTargets {
			buildNormal(build.Target{Platform: build.TargetPlatforms.Linux}, nil)
		}
		if hasDarwinTargets {
			buildNormal(build.Target{Platform: build.TargetPlatforms.Darwin}, nil)
		}
		if hasWindowsTargets {
			buildNormal(build.Target{Platform: build.TargetPlatforms.Windows}, nil)
		}
		for _, buildTarget := range buildTargets {
			if buildTarget.PackagingFormat == build.TargetPackagingFormats.AppImage {
				packaging.BuildLinuxAppImage(buildTarget)
			} else if buildTarget.PackagingFormat == build.TargetPackagingFormats.Deb {
				packaging.BuildLinuxDeb(buildTarget)
			} else if buildTarget.PackagingFormat == build.TargetPackagingFormats.Snap {
				packaging.BuildLinuxSnap(buildTarget)
			} else if buildTarget.PackagingFormat == build.TargetPackagingFormats.Bundle {
				packaging.BuildDarwinBundle(buildTarget)
			} else if buildTarget.PackagingFormat == build.TargetPackagingFormats.Pkg {
				packaging.BuildDarwinPkg(buildTarget)
			} else if buildTarget.PackagingFormat == build.TargetPackagingFormats.Msi {
				packaging.BuildWindowsMsi(buildTarget)
			}
		}
	},
}

// checkForMainDesktop checks and adds the lib/main_desktop.dart dart entry
// point if needed
func checkForMainDesktop() {
	if buildTargetFile != "lib/main_desktop.dart" {
		return
	}
	_, err := os.Stat("lib/main_desktop.dart")
	if os.IsNotExist(err) {
		log.Warnf("target file \"lib/main_desktop.dart\" not found.")
		log.Warnf("Let hover add the \"lib/main_desktop.dart\" file? ")
		if askForConfirmation() {
			fileutils.CopyAsset("app/main_desktop.dart", filepath.Join("lib", "main_desktop.dart"), fileutils.AssetsBox)
			log.Infof("target file \"lib/main_desktop.dart\" has been created.")
			log.Infof("       Depending on your project, you might want to tweak it.")
			return
		}
		log.Printf("You can define a custom traget by using the %s flag.", log.Au().Magenta("--target"))
		os.Exit(1)
	}
	if err != nil {
		log.Errorf("Failed to stat lib/main_desktop.dart: %v", err)
		os.Exit(1)
	}
}

func overrideBuildConfig() {
	if buildTargetFile == config.BuildTargetFileDefault && config.GetConfig().Target != "" {
		buildTargetFile = config.GetConfig().Target
	}
	if buildBranch == config.BuildBranchDefault && config.GetConfig().Branch != "" {
		buildBranch = config.GetConfig().Branch
	}
	if buildCachePath == config.BuildCachePathDefault && config.GetConfig().CachePath != "" {
		buildCachePath = config.GetConfig().CachePath
	}
	if buildOpenGlVersion == config.BuildOpenGlVersionDefault && config.GetConfig().OpenGL != "" {
		buildOpenGlVersion = config.GetConfig().OpenGL
	}
	if !buildDocker && config.GetConfig().Docker {
		buildDocker = config.GetConfig().Docker
	}
}

func runPluginGet() {
	// must be run before `flutter build bundle`
	// because `build bundle` will update the file timestamp
	runPluginGet, err := shouldRunPluginGet()
	if err != nil {
		log.Errorf("Failed to check if plugin get should be run: %v.", err)
		os.Exit(1)
	}

	if runPluginGet {
		log.Printf("listing available plugins:")
		if hoverPluginGet(true) {
			log.Infof(fmt.Sprintf("run `%s`? ", log.Au().Magenta("hover plugins get")))
			if askForConfirmation() {
				hoverPluginGet(false)
			}
		}
	}
}

func buildFlutterBundle() {
	var trackWidgetCreation string
	if buildDebug {
		trackWidgetCreation = "--track-widget-creation"
	}
	cmdFlutterBuild := exec.Command(build.FlutterBin, "build", "bundle",
		"--asset-dir", build.BundlePath,
		"--target", buildTargetFile,
		trackWidgetCreation,
	)
	cmdFlutterBuild.Stderr = os.Stderr
	cmdFlutterBuild.Stdout = os.Stdout

	if !buildOmitFlutterBundle {
		log.Infof("Bundling flutter app")
		err := cmdFlutterBuild.Run()
		if err != nil {
			log.Errorf("Flutter build failed: %v", err)
			os.Exit(1)
		}
	}
}

func checkFlutter() {
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
				log.Warnf("⚠     It`s advised to use the beta channel: `%s`", log.Au().Magenta("flutter channel beta"))
			}
		} else {
			log.Warnf("Failed to check your flutter channel: Unrecognized output format")
		}
	}
}

func buildInDocker(buildTarget build.Target, vmArguments []string) {
	crossCompilingDir, _ := filepath.Abs(filepath.Join(build.BuildPath, "cross-compiling"))
	err := os.MkdirAll(crossCompilingDir, 0755)
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
			// TODO: Once golang 1.13 is available in the 'dockercore/golang-cross' image.
			// Use the 'dockercore' image instead. Pending PR: https://github.com/docker/golang-cross/pull/45
			"FROM goreng/golang-cross",
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
		log.Infof("A Dockerfile for cross-compiling for %s has been created at %s. You can add it to git.", buildTarget.Platform, dockerFilePath)
	}
	log.Printf("Building docker image")
	dockerBuildCmd := exec.Command(build.DockerBin, "build", "-t", "hover-build-cc", ".")
	dockerBuildCmd.Dir = crossCompilingDir
	err = dockerBuildCmd.Run()
	if err != nil {
		log.Errorf("Docker build failed: %v", err)
		os.Exit(1)
	}

	log.Infof("Cross-Compiling `go-flutter` and plugins using docker")

	u, err := user.Current()
	if err != nil {
		log.Errorf("Couldn`t get current user: %v", err)
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
	for _, env := range buildEnv(buildTarget, "/engine") {
		args = append(args, "-e", env)
	}
	args = append(args, "hover-build-cc")
	chownStr := ""
	if runtime.GOOS != "windows" {
		chownStr = fmt.Sprintf(" && chown %s:%s ./ -R", u.Uid, u.Gid)
	}
	stripStr := ""
	if buildTarget.Platform == build.TargetPlatforms.Linux {
		outputEngineFile := filepath.Join("/app", build.BuildPath, "build", "outputs", buildTarget.Platform, build.EngineFile(buildTarget))
		stripStr = fmt.Sprintf("strip -s %s && ", outputEngineFile)
	}
	args = append(args, "bash", "-c", fmt.Sprintf("%s%s%s", stripStr, strings.Join(buildCommand(buildTarget, vmArguments, "build/outputs/"+buildTarget.Platform+"/"+build.OutputBinaryName(pubspec.GetPubSpec().Name, buildTarget)), " "), chownStr))
	dockerRunCmd := exec.Command(build.DockerBin, args...)
	dockerRunCmd.Stderr = os.Stderr
	dockerRunCmd.Stdout = os.Stdout
	dockerRunCmd.Dir = crossCompilingDir
	err = dockerRunCmd.Run()
	if err != nil {
		log.Errorf("Docker run failed: %v", err)
		os.Exit(1)
	}
	log.Infof("Successfully cross-compiled for " + buildTarget.Platform)
}

func buildNormal(buildTarget build.Target, vmArguments []string) {
	crossCompile = buildTarget.Platform != runtime.GOOS
	buildDocker = crossCompile || buildDocker

	if buildCachePath != "" {
		engineCachePath = enginecache.ValidateOrUpdateEngineAtPath(buildTarget, buildCachePath)
	} else {
		engineCachePath = enginecache.ValidateOrUpdateEngine(buildTarget)
	}

	if !buildOmitFlutterBundle && !buildOmitEmbedder {
		err := os.RemoveAll(build.OutputDirectoryPath(buildTarget, false))
		log.Printf("Cleaning the build directory")
		if err != nil {
			log.Errorf("Failed to clean output directory %s: %v", build.OutputDirectoryPath(buildTarget, false), err)
			os.Exit(1)
		}
	}
	fileutils.CopyDir(build.IntermediatesDirectoryPath(targetOS), build.OutputDirectoryPath(targetOS))

	err := os.MkdirAll(build.OutputDirectoryPath(buildTarget, false), 0775)
	if err != nil {
		log.Errorf("Failed to create output directory %s: %v", build.OutputDirectoryPath(buildTarget, false), err)
		os.Exit(1)
	}

	err = copy.Copy(
		build.BundlePath,
		filepath.Join(build.OutputDirectoryPath(buildTarget, false), "flutter_assets"),
	)
	if err != nil {
		log.Errorf("Failed to copy bundle: %v", err)
		os.Exit(1)
	}
	outputEngineFile := filepath.Join(build.OutputDirectoryPath(buildTarget, false), build.EngineFile(buildTarget))
	err = copy.Copy(
		filepath.Join(engineCachePath, build.EngineFile(buildTarget)),
		outputEngineFile,
	)
	if err != nil {
		log.Errorf("Failed to copy %s: %v", build.EngineFile(buildTarget), err)
		os.Exit(1)
	}

	err = copy.Copy(
		filepath.Join(engineCachePath, "artifacts", "icudtl.dat"),
		filepath.Join(build.OutputDirectoryPath(buildTarget, false), "icudtl.dat"),
	)
	if err != nil {
		log.Errorf("Failed to copy icudtl.dat: %v", err)
		os.Exit(1)
	}

	err = copy.Copy(
		filepath.Join(build.BuildPath, "assets"),
		filepath.Join(build.OutputDirectoryPath(buildTarget, false), "assets"),
	)
	if err != nil {
		log.Errorf("Failed to copy %s/assets: %v", build.BuildPath, err)
		os.Exit(1)
	}

	if buildOmitEmbedder {
		// Omit the `go-flutter` build
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
			log.Errorf("Faild to parse `go-flutter` semver: %v", err)
			os.Exit(1)
		}

		if semver.Prerelease() != "" {
			log.Infof("Upgrading `go-flutter` to the latest release")
			// no buildBranch provided and currentTag isn`t a release,
			// force update. (same behaviour as previous version of hover).
			err = upgradeGoFlutter(buildTarget, engineCachePath)
			if err != nil {
				// the upgrade can fail silently
				log.Warnf("Upgrade ignored, current `go-flutter` version: %s", currentTag)
			}
		} else {
			// when the buildBranch is empty and the currentTag is a release.
			// Check if the `go-flutter` needs updates.
			versioncheck.CheckForGoFlutterUpdate(filepath.Join(wd, build.BuildPath), currentTag)
		}

	} else {
		log.Printf("Downloading `go-flutter` %s", buildBranch)

		// when the buildBranch is set, fetch the go-flutter branch version.
		err = upgradeGoFlutter(buildTarget, engineCachePath)
		if err != nil {
			os.Exit(1)
		}
	}

	if buildOpenGlVersion == "none" {
		log.Warnf("The `--opengl=none` flag makes go-flutter incompatible with texture plugins!")
	}

	if buildDocker {
		if crossCompile {
			log.Infof("Because %s is not able to compile for %s out of the box, a cross-compiling container is used", runtime.GOOS, buildTarget.Platform)
		}
		buildInDocker(buildTarget, vmArguments)
		return
	}

	if !buildDebug && buildTarget.Platform == build.TargetPlatforms.Linux {
		err = exec.Command("strip", "-s", outputEngineFile).Run()
		if err != nil {
			log.Errorf("Failed to strip %s: %v", outputEngineFile, err)
			os.Exit(1)
		}
	}

	buildCommandString := buildCommand(buildTarget, vmArguments, build.OutputBinaryPath(pubspec.GetPubSpec().Name, buildTarget, false))
	cmdGoBuild := exec.Command(buildCommandString[0], buildCommandString[1:]...)
	cmdGoBuild.Dir = filepath.Join(wd, build.BuildPath)
	cmdGoBuild.Env = append(os.Environ(),
		buildEnv(buildTarget, engineCachePath)...,
	)

	cmdGoBuild.Stderr = os.Stderr
	cmdGoBuild.Stdout = os.Stdout

	log.Infof("Compiling `go-flutter` and plugins")
	err = cmdGoBuild.Run()
	if err != nil {
		log.Errorf("Go build failed: %v", err)
		os.Exit(1)
	}
	log.Infof("Successfully compiled")
}

func buildEnv(buildTarget build.Target, engineCachePath string) []string {
	env := []string{
		"GO111MODULE=on",
		"CGO_LDFLAGS=" + build.CGoLdFlags(buildTarget, engineCachePath),
		"GOOS=" + buildTarget.Platform,
		"GOARCH=amd64",
		"CGO_ENABLED=1",
	}
	if buildDocker {
		env = append(env,
			"GOCACHE=/cache",
		)
		if buildTarget.Platform == build.TargetPlatforms.Windows {
			env = append(env,
				"CC="+mingwGccBinName,
			)
		}
		if buildTarget.Platform == build.TargetPlatforms.Darwin {
			env = append(env,
				"CC="+clangBinName,
			)
		}
	}
	return env
}

func buildCommand(buildTarget build.Target, vmArguments []string, outputBinaryPath string) []string {
	currentTag, err := versioncheck.CurrentGoFlutterTag(build.BuildPath)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

	var ldflags []string
	if !buildDebug {
		vmArguments = append(vmArguments, "--disable-dart-asserts")
		vmArguments = append(vmArguments, "--disable-observatory")

		if buildTarget.Platform == build.TargetPlatforms.Windows {
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
		androidmanifest.AndroidOrganizationName()))

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
