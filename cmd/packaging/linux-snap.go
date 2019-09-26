package packaging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

func InitLinuxSnap() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-snap"
	createPackagingFormatDirectory(packagingFormat)
	snapDirectoryPath := packagingFormatPath(packagingFormat)

	snapLocalDirectoryPath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "snap", "local"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for snap local directory: %v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(snapLocalDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create snap local directory %s: %v\n", snapDirectoryPath, err)
		os.Exit(1)
	}

	snapcraftFilePath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "snap", "snapcraft.yaml"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for snapcraft.yaml file %s: %v\n", snapcraftFilePath, err)
		os.Exit(1)
	}

	snapcraftFile, err := os.Create(snapcraftFilePath)
	if err != nil {
		fmt.Printf("hover: Failed to create snapcraft.yaml file %s: %v\n", snapcraftFilePath, err)
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
			fmt.Printf("hover: Could not write snapcraft.yaml: %v\n", err)
			os.Exit(1)
		}
	}
	err = snapcraftFile.Close()
	if err != nil {
		fmt.Printf("hover: Could not close snapcraft.yaml: %v\n", err)
		os.Exit(1)
	}

	desktopFilePath, err := filepath.Abs(filepath.Join(snapLocalDirectoryPath, projectName+".desktop"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for desktop file %s: %v\n", desktopFilePath, err)
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
	fmt.Printf("hover: Packaging snap in %s\n", tmpPath)

	err := copy.Copy(filepath.Join(build.BuildPath, "assets"), filepath.Join(tmpPath, "assets"))
	if err != nil {
		fmt.Printf("hover: Could not copy assets folder: %v\n", err)
		os.Exit(1)
	}
	err = copy.Copy(build.OutputDirectoryPath("linux"), filepath.Join(tmpPath, "build"))
	if err != nil {
		fmt.Printf("hover: Could not copy build folder: %v\n", err)
		os.Exit(1)
	}
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		fmt.Printf("hover: Could not copy packaging configuration folder: %v\n", err)
		os.Exit(1)
	}

	outputFileName := removeDashesAndUnderscores(projectName) + "_" + pubspec.GetPubSpec().Version + "_" + runtime.GOARCH + ".snap"
	outputFilePath := filepath.Join(build.OutputDirectoryPath("linux-snap"), outputFileName)
	runDockerPackaging(tmpPath, packagingFormat, []string{"snapcraft"})

	err = os.Rename(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		fmt.Printf("hover: Could not move snap file: %v\n", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpPath)
	if err != nil {
		fmt.Printf("hover: Could not remove temporary build directory: %v\n", err)
		os.Exit(1)
	}
	printPackagingFinished(packagingFormat)
}
