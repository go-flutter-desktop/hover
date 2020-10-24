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
func buildDirectoryPath(targetOS string, mode Mode, path string) string {
	outputDirectoryPath, err := filepath.Abs(filepath.Join(BuildPath, "build", path, fmt.Sprintf("%s-%s", targetOS, mode.Name)))
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
func OutputDirectoryPath(targetOS string, mode Mode) string {
	return buildDirectoryPath(targetOS, mode, "outputs")
}

// IntermediatesDirectoryPath returns the path where the intermediates stored.
// If needed, the directory is create at the returned path.
//
// Those intermediates include the dynamic library dependencies of go-flutter plugins.
// hover copies these intermediates from flutter plugins folder when `hover plugins get`, and
// copies to go-flutter's binary output folder before build.
func IntermediatesDirectoryPath(targetOS string, mode Mode) string {
	return buildDirectoryPath(targetOS, mode, "intermediates")
}

// OutputBinary returns the string of the executable used to launch the
// main desktop app. (appends .exe for windows)
func OutputBinary(executableName, targetOS string) string {
	var outputBinaryName = executableName
	switch targetOS {
	case "darwin":
		// no special filename
	case "linux":
		// no special filename
	case "windows":
		outputBinaryName += ".exe"
	default:
		log.Errorf("Target platform %s is not supported.", targetOS)
		os.Exit(1)
	}
	return outputBinaryName
}

// OutputBinaryPath returns the path to the go-flutter Application for a
// specified platform.
func OutputBinaryPath(executableName, targetOS string, mode Mode) string {
	outputBinaryPath := filepath.Join(OutputDirectoryPath(targetOS, mode), OutputBinary(executableName, targetOS))
	return outputBinaryPath
}

// ExecutableExtension returns the extension of binary files on a given platform
func ExecutableExtension(targetOS string) string {
	switch targetOS {
	case "darwin":
		// no special filename
		return ""
	case "linux":
		// no special filename
		return ""
	case "windows":
		return ".exe"
	default:
		log.Errorf("Target platform %s is not supported.", targetOS)
		os.Exit(1)
		return ""
	}
}

// EngineFiles returns the names of the engine files from flutter for the
// specified platform and build mode.
func EngineFiles(targetOS string, mode Mode) []string {
	switch targetOS {
	case "darwin":
		if mode.IsAot {
			return []string{"libflutter_engine.dylib"}
		} else {
			return []string{"FlutterEmbedder.framework"}
		}
	case "linux":
		return []string{"libflutter_engine.so"}
	case "windows":
		if mode.IsAot {
			return []string{"flutter_engine.dll", "flutter_engine.exp", "flutter_engine.lib", "flutter_engine.pdb"}
		} else {
			return []string{"flutter_engine.dll"}
		}
	default:
		log.Errorf("%s has no implemented engine file", targetOS)
		os.Exit(1)
		return []string{}
	}
}
