package packaging

import (
	"os"
	"path/filepath"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

func InitLinuxAppImage() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-appimage"
	createPackagingFormatDirectory(packagingFormat)
	appImageDirectoryPath := packagingFormatPath(packagingFormat)
	appRunFilePath, err := filepath.Abs(filepath.Join(appImageDirectoryPath, "AppRun"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for AppRun file %s: %v", appRunFilePath, err)
		os.Exit(1)
	}

	appRunFile, err := os.Create(appRunFilePath)
	if err != nil {
		log.Errorf("Failed to create AppRun file %s: %v", appRunFilePath, err)
		os.Exit(1)
	}
	appRunFileContent := []string{
		`#!/bin/sh`,
		`cd "$(dirname "$0")"`,
		`exec ./build/` + projectName,
	}
	for _, line := range appRunFileContent {
		if _, err := appRunFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write AppRun file: %v", err)
			os.Exit(1)
		}
	}
	err = appRunFile.Close()
	if err != nil {
		log.Errorf("Could not close AppRun file: %v", err)
		os.Exit(1)
	}
	err = os.Chmod(appRunFilePath, 0777)
	if err != nil {
		log.Errorf("Failed to change file permissions for AppRun file: %v", err)
		os.Exit(1)
	}

	desktopFilePath, err := filepath.Abs(filepath.Join(appImageDirectoryPath, projectName+".desktop"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for desktop file %s: %v", desktopFilePath, err)
		os.Exit(1)
	}
	createLinuxDesktopFile(desktopFilePath, packagingFormat, "", "/build/assets/icon")
	createDockerfile(packagingFormat)

	printInitFinished(packagingFormat)
}

func BuildLinuxAppImage() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-appimage"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	log.Infof("Packaging AppImage in %s", tmpPath)

	err := copy.Copy(build.OutputDirectoryPath("linux"), filepath.Join(tmpPath, "build"))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	outputFileName := projectName + "-x86_64.AppImage"
	outputFilePath := filepath.Join(build.OutputDirectoryPath("linux-appimage"), outputFileName)
	runDockerPackaging(tmpPath, packagingFormat, []string{"appimagetool", "."})

	err = os.Rename(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move AppImage file: %v", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpPath)
	if err != nil {
		log.Errorf("Could not remove temporary build directory: %v", err)
		os.Exit(1)
	}
	printPackagingFinished(packagingFormat)
}
