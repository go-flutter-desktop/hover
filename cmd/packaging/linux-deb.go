package packaging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

func InitLinuxDeb() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-deb"
	createPackagingFormatDirectory(packagingFormat)
	debDirectoryPath := packagingFormatPath(packagingFormat)
	debDebianDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "DEBIAN"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for DEBIAN directory: %v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(debDebianDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create DEBIAN directory %s: %v\n", debDebianDirectoryPath, err)
		os.Exit(1)
	}

	binDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "usr", "bin"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for bin directory: %v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(binDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create bin directory %s: %v\n", binDirectoryPath, err)
		os.Exit(1)
	}
	applicationsDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "usr", "share", "applications"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for applications directory: %v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(applicationsDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create applications directory %s: %v\n", applicationsDirectoryPath, err)
		os.Exit(1)
	}

	controlFilePath, err := filepath.Abs(filepath.Join(debDebianDirectoryPath, "control"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for control file %s: %v\n", controlFilePath, err)
		os.Exit(1)
	}

	controlFile, err := os.Create(controlFilePath)
	if err != nil {
		fmt.Printf("hover: Failed to create control file %s: %v\n", controlFilePath, err)
		os.Exit(1)
	}
	controlFileContent := []string{
		"Package: " + removeDashesAndUnderscores(projectName),
		"Architecture: amd64",
		"Maintainer: @" + getAuthor(),
		"Priority: optional",
		"Version: " + pubspec.GetPubSpec().Version,
		"Description: " + pubspec.GetPubSpec().Description,
		"Depends: " + strings.Join(linuxPackagingDependencies, ","),
	}

	for _, line := range controlFileContent {
		if _, err := controlFile.WriteString(line + "\n"); err != nil {
			fmt.Printf("hover: Could not write control file: %v\n", err)
			os.Exit(1)
		}
	}
	err = controlFile.Close()
	if err != nil {
		fmt.Printf("hover: Could not close control file: %v\n", err)
		os.Exit(1)
	}

	binFilePath, err := filepath.Abs(filepath.Join(binDirectoryPath, removeDashesAndUnderscores(projectName)))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for bin file %s: %v\n", binFilePath, err)
		os.Exit(1)
	}

	binFile, err := os.Create(binFilePath)
	if err != nil {
		fmt.Printf("hover: Failed to create bin file %s: %v\n", controlFilePath, err)
		os.Exit(1)
	}
	binFileContent := []string{
		"#!/bin/sh",
		"/usr/lib/" + projectName + "/" + projectName,
	}
	for _, line := range binFileContent {
		if _, err := binFile.WriteString(line + "\n"); err != nil {
			fmt.Printf("hover: Could not write bin file: %v\n", err)
			os.Exit(1)
		}
	}
	err = binFile.Close()
	if err != nil {
		fmt.Printf("hover: Could not close bin file: %v\n", err)
		os.Exit(1)
	}
	err = os.Chmod(binFilePath, 0777)
	if err != nil {
		fmt.Printf("hover: Failed to change file permissions for bin file: %v\n", err)
		os.Exit(1)
	}

	desktopFilePath, err := filepath.Abs(filepath.Join(applicationsDirectoryPath, projectName+".desktop"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for desktop file %s: %v\n", desktopFilePath, err)
		os.Exit(1)
	}
	createLinuxDesktopFile(desktopFilePath, packagingFormat, "/usr/bin/"+projectName, "/usr/lib/"+projectName+"/assets/icon.png")
	createDockerfile(packagingFormat)

	printInitFinished(packagingFormat)
}

func BuildLinuxDeb() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-deb"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	fmt.Printf("hover: Packaging deb in %s\n", tmpPath)

	libDirectoryPath, err := filepath.Abs(filepath.Join(tmpPath, "usr", "lib"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for lib directory: %v\n", err)
		os.Exit(1)
	}
	err = copy.Copy(build.OutputDirectoryPath("linux"), filepath.Join(libDirectoryPath, projectName))
	if err != nil {
		fmt.Printf("hover: Could not copy build folder: %v\n", err)
		os.Exit(1)
	}
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		fmt.Printf("hover: Could not copy packaging configuration folder: %v\n", err)
		os.Exit(1)
	}

	outputFileName := removeDashesAndUnderscores(projectName) + "_" + runtime.GOARCH + ".deb"
	outputFilePath := filepath.Join(build.OutputDirectoryPath("linux-deb"), outputFileName)
	runDockerPackaging(tmpPath, packagingFormat, []string{"dpkg-deb", "--build", ".", outputFileName})

	err = os.Rename(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		fmt.Printf("hover: Could not move deb file: %v\n", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpPath)
	if err != nil {
		fmt.Printf("hover: Could not remove temporary build directory: %v\n", err)
		os.Exit(1)
	}
	printPackagingFinished(packagingFormat)
}
