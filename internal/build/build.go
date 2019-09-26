package build

import (
	"github.com/go-flutter-desktop/hover/internal/log"
	"os"
	"path/filepath"
)

const BuildPath = "go"

func OutputDirectoryPath(targetOS string) string {
	outputDirectoryPath, err := filepath.Abs(filepath.Join(BuildPath, "build", "outputs", targetOS))
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

func OutputBinaryName(projectName string, targetOS string) string {
	var outputBinaryName = projectName
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

func OutputBinaryPath(projectName string, targetOS string) string {
	outputBinaryPath := filepath.Join(OutputDirectoryPath(targetOS), OutputBinaryName(projectName, targetOS))
	return outputBinaryPath
}
