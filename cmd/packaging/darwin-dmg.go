package packaging

import (
	"os"
	"path/filepath"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

// InitDarwinDmg initializes the a darwin dmg packaging format.
func InitDarwinDmg() {
	packagingFormat := "darwin-dmg"
	if _, err := os.Stat(packagingFormatPath("darwin-bundle")); os.IsNotExist(err) {
		log.Errorf("`darwin-dmg` depends on `darwin-bundle`. Run `hover init-packaging darwin-bundle` first and then `hover init-packaging darwin-dmg`")
		os.Exit(1)
	}
	createPackagingFormatDirectory(packagingFormat)

	createDockerfile(packagingFormat, []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install genisoimage -y ",
	})

	printInitFinished(packagingFormat)
}

// BuildDarwinDmg uses the InitDarwinDmg template to create an dmg package.
func BuildDarwinDmg(buildVersion string) {
	log.Infof("Building darwin-bundle first")
	BuildDarwinBundle(buildVersion)
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "darwin-dmg"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
	log.Infof("Packaging dmg in %s", tmpPath)

	err := copy.Copy(build.OutputDirectoryPath("darwin-bundle"), filepath.Join(tmpPath, "dmgdir"))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	outputFileName := projectName + "-" + buildVersion + ".dmg"
	runDockerPackaging(tmpPath, packagingFormat, []string{
		"genisoimage -V '" + projectName + "' -D -R -apple -no-pad -o '" + outputFileName + "' dmgdir",
	})

	outputFilePath := filepath.Join(build.OutputDirectoryPath("darwin-dmg"), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move dmg: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(packagingFormat)
}
