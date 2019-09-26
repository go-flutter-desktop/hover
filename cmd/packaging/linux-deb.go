package packaging

import (
	"github.com/go-flutter-desktop/hover/internal/log"
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
		log.Errorf("Failed to resolve absolute path for DEBIAN directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(debDebianDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create DEBIAN directory %s: %v", debDebianDirectoryPath, err)
		os.Exit(1)
	}

	binDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "usr", "bin"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for bin directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(binDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create bin directory %s: %v", binDirectoryPath, err)
		os.Exit(1)
	}
	applicationsDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "usr", "share", "applications"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for applications directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(applicationsDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create applications directory %s: %v", applicationsDirectoryPath, err)
		os.Exit(1)
	}

	controlFilePath, err := filepath.Abs(filepath.Join(debDebianDirectoryPath, "control"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for control file %s: %v", controlFilePath, err)
		os.Exit(1)
	}

	controlFile, err := os.Create(controlFilePath)
	if err != nil {
		log.Errorf("Failed to create control file %s: %v", controlFilePath, err)
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
			log.Errorf("Could not write control file: %v", err)
			os.Exit(1)
		}
	}
	err = controlFile.Close()
	if err != nil {
		log.Errorf("Could not close control file: %v", err)
		os.Exit(1)
	}

	binFilePath, err := filepath.Abs(filepath.Join(binDirectoryPath, removeDashesAndUnderscores(projectName)))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for bin file %s: %v", binFilePath, err)
		os.Exit(1)
	}

	binFile, err := os.Create(binFilePath)
	if err != nil {
		log.Errorf("Failed to create bin file %s: %v", controlFilePath, err)
		os.Exit(1)
	}
	binFileContent := []string{
		"#!/bin/sh",
		"/usr/lib/" + projectName + "/" + projectName,
	}
	for _, line := range binFileContent {
		if _, err := binFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write bin file: %v", err)
			os.Exit(1)
		}
	}
	err = binFile.Close()
	if err != nil {
		log.Errorf("Could not close bin file: %v", err)
		os.Exit(1)
	}
	err = os.Chmod(binFilePath, 0777)
	if err != nil {
		log.Errorf("Failed to change file permissions for bin file: %v", err)
		os.Exit(1)
	}

	desktopFilePath, err := filepath.Abs(filepath.Join(applicationsDirectoryPath, projectName+".desktop"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for desktop file %s: %v", desktopFilePath, err)
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
	log.Infof("Packaging deb in %s", tmpPath)

	libDirectoryPath, err := filepath.Abs(filepath.Join(tmpPath, "usr", "lib"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for lib directory: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(build.OutputDirectoryPath("linux"), filepath.Join(libDirectoryPath, projectName))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	outputFileName := removeDashesAndUnderscores(projectName) + "_" + runtime.GOARCH + ".deb"
	outputFilePath := filepath.Join(build.OutputDirectoryPath("linux-deb"), outputFileName)
	runDockerPackaging(tmpPath, packagingFormat, []string{"dpkg-deb", "--build", ".", outputFileName})

	err = os.Rename(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move deb file: %v", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpPath)
	if err != nil {
		log.Errorf("Could not remove temporary build directory: %v", err)
		os.Exit(1)
	}
	printPackagingFinished(packagingFormat)
}
