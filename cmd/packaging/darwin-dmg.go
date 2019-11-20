package packaging

import (
	"os"
	"path/filepath"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

func InitDarwinDmg(buildTarget build.Target) {
	if _, err := os.Stat(PackagingFormatPath(buildTarget)); os.IsNotExist(err) {
		log.Errorf("`darwin-dmg` depends on `darwin-bundle`. Run `hover init-packaging darwin-bundle` first and then `hover init-packaging darwin-dmg`")
		os.Exit(1)
	}
	createPackagingFormatDirectory(buildTarget)

	createDockerfile(buildTarget, []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install genisoimage -y ",
	})

	printInitFinished(buildTarget)
}

func BuildDarwinDmg(buildTarget build.Target) {
	projectName := pubspec.GetPubSpec().Name
	tmpPath := getTemporaryBuildDirectory(projectName, buildTarget)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
	log.Infof("Packaging dmg in %s", tmpPath)

	err := copy.Copy(build.OutputDirectoryPath(buildTarget, false), filepath.Join(tmpPath, "dmgdir"))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(PackagingFormatPath(buildTarget), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	outputFileName := projectName + " " + pubspec.GetPubSpec().Version + ".dmg"
	runDockerPackaging(tmpPath, buildTarget, []string{
		"genisoimage -V '" + projectName + "' -D -R -apple -no-pad -o '" + outputFileName + "' dmgdir",
	})

	outputFilePath := filepath.Join(build.OutputDirectoryPath(buildTarget, true), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move dmg: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(buildTarget)
}
