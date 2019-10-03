package packaging

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
	"github.com/go-flutter-desktop/hover/internal/log"
)

func InitLinuxSnap() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-snap"
	createPackagingFormatDirectory(packagingFormat)
	snapDirectoryPath := packagingFormatPath(packagingFormat)

	snapLocalDirectoryPath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "snap", "local"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for snap local directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(snapLocalDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create snap local directory %s: %v", snapDirectoryPath, err)
		os.Exit(1)
	}

	snapcraftFilePath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "snap", "snapcraft.yaml"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for snapcraft.yaml file %s: %v", snapcraftFilePath, err)
		os.Exit(1)
	}

	snapcraftFile, err := os.Create(snapcraftFilePath)
	if err != nil {
		log.Errorf("Failed to create snapcraft.yaml file %s: %v", snapcraftFilePath, err)
		os.Exit(1)
	}
	snapcraftFileContent := []string{
		"name: " + removeDashesAndUnderscores(projectName),
		"base: core18",
		"version: '" + pubspec.GetPubSpec().Version + "'",
		"summary: " + pubspec.GetPubSpec().Description,
		"description: |",
		"  " + pubspec.GetPubSpec().Description,
		"confinement: devmode",
		"grade: devel",
		"apps:",
		"  " + removeDashesAndUnderscores(projectName) + ":",
		"    command: " + projectName,
		"    desktop: local/" + projectName + ".desktop",
		"parts:",
		"  desktop:",
		"    plugin: dump",
		"    source: snap",
		"  assets:",
		"    plugin: dump",
		"    source: assets",
		"  app:",
		"    plugin: dump",
		"    source: build",
		"    stage-packages:",
	}
	for _, dependency := range linuxPackagingDependencies {
		snapcraftFileContent = append(snapcraftFileContent, "      - "+dependency)
	}

	for _, line := range snapcraftFileContent {
		if _, err := snapcraftFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write snapcraft.yaml: %v", err)
			os.Exit(1)
		}
	}
	err = snapcraftFile.Close()
	if err != nil {
		log.Errorf("Could not close snapcraft.yaml: %v", err)
		os.Exit(1)
	}

	desktopFilePath, err := filepath.Abs(filepath.Join(snapLocalDirectoryPath, projectName+".desktop"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for desktop file %s: %v", desktopFilePath, err)
		os.Exit(1)
	}
	createLinuxDesktopFile(desktopFilePath, packagingFormat, "/"+projectName, "/icon.png")
	createDockerfile(packagingFormat)

	printInitFinished(packagingFormat)
}

func BuildLinuxSnap() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-snap"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	log.Infof("Packaging snap in %s", tmpPath)

	err := copy.Copy(filepath.Join(build.BuildPath, "assets"), filepath.Join(tmpPath, "assets"))
	if err != nil {
		log.Errorf("Could not copy assets folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(build.OutputDirectoryPath("linux"), filepath.Join(tmpPath, "build"))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	outputFileName := removeDashesAndUnderscores(projectName) + "_" + pubspec.GetPubSpec().Version + "_" + runtime.GOARCH + ".snap"
	outputFilePath := filepath.Join(build.OutputDirectoryPath("linux-snap"), outputFileName)
	runDockerPackaging(tmpPath, packagingFormat, []string{"snapcraft"})

	err = os.Rename(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move snap file: %v", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpPath)
	if err != nil {
		log.Errorf("Could not remove temporary build directory: %v", err)
		os.Exit(1)
	}
	printPackagingFinished(packagingFormat)
}
