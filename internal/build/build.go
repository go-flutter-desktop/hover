package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-flutter-desktop/hover/internal/log"
)

// BuildPath sets the name of the directory used to store the go-flutter project.
// Much like android and ios are already used.
const BuildPath = "go"

// buildDirectoryPath returns the path in `BuildPath`/build.
// If needed, the directory is create at the returned path.
func buildDirectoryPath(buildTarget Target, withPackagingFormat bool, path string) string {
	targetName := buildTarget.Platform
	if buildTarget.PackagingFormat != "" && withPackagingFormat {
		targetName += fmt.Sprintf("-%s", buildTarget.PackagingFormat)
	}
	outputDirectoryPath, err := filepath.Abs(filepath.Join(BuildPath, "build", path, targetName))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for output directory: %v", err)
		os.Exit(1)
	}
	if _, err := os.Stat(outputDirectoryPath); os.IsNotExist(err) {
		err = os.MkdirAll(outputDirectoryPath, 0775)
		if err != nil {
			log.Errorf("Failed to create output directory %s: %v", outputDirectoryPath, err)
			os.Exit(1)
		}
	}
	return outputDirectoryPath
}

// OutputDirectoryPath returns the path where the go-flutter binary and flutter
// binaries blobs will be stored for a particular platform.
// If needed, the directory is create at the returned path.
func OutputDirectoryPath(buildTarget Target, withPackagingFormat bool) string {
	return buildDirectoryPath(targetOS, "outputs")
}

// IntermediatesDirectoryPath returns the path where the intermediates stored.
// If needed, the directory is create at the returned path.
//
// Those intermediates include the dynamic library dependencies of go-flutter plugins.
// hover copies these intermediates from flutter plugins folder when `hover plugins get`, and
// copies to go-flutter's binary output folder before build.
func IntermediatesDirectoryPath(targetOS string) string {
	return buildDirectoryPath(targetOS, "intermediates")
}

// OutputBinaryName returns the string of the executable used to launch the
// main desktop app. (appends .exe for windows)
func OutputBinaryName(projectName string, buildTarget Target) string {
	var outputBinaryName = projectName
	switch buildTarget.Platform {
	case TargetPlatforms.Darwin:
		// no special filename
	case TargetPlatforms.Linux:
		// no special filename
	case TargetPlatforms.Windows:
		outputBinaryName += ".exe"
	default:
		log.Errorf("target platform %s is not supported.", buildTarget.Platform)
		os.Exit(1)
	}
	return outputBinaryName
}

// OutputBinaryPath returns the path to the go-flutter Application for a
// specified platform.
func OutputBinaryPath(projectName string, buildTarget Target, withPackagingFormat bool) string {
	outputBinaryPath := filepath.Join(OutputDirectoryPath(buildTarget, withPackagingFormat), OutputBinaryName(projectName, buildTarget))
	return outputBinaryPath
}

// EngineFile returns the name of the engine file from flutter for the
// specified platform.
func EngineFile(buildTarget Target) string {
	switch buildTarget.Platform {
	case TargetPlatforms.Darwin:
		return "FlutterEmbedder.framework"
	case TargetPlatforms.Linux:
		return "libflutter_engine.so"
	case TargetPlatforms.Windows:
		return "flutter_engine.dll"
	default:
		log.Errorf("%s has no implemented engine file", buildTarget.Platform)
		os.Exit(1)
		return ""
	}
}

func CGoLdFlags(buildTarget Target, engineCachePath string) string {
	var cgoLdflags string
	switch buildTarget.Platform {
	case TargetPlatforms.Darwin:
		cgoLdflags = fmt.Sprintf("-F%s -Wl,-rpath,@executable_path", engineCachePath)
	case TargetPlatforms.Linux:
		cgoLdflags = fmt.Sprintf("-L%s", engineCachePath)
	case TargetPlatforms.Windows:
		cgoLdflags = fmt.Sprintf("-L%s", engineCachePath)
	default:
		log.Errorf("target platform %s is not supported, cgo_ldflags not implemented.", buildTarget.Platform)
		os.Exit(1)
	}
	return cgoLdflags
}
