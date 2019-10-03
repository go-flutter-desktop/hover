package packaging

import (
	"github.com/otiai10/copy"
	"os"
	"path/filepath"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

func InitDarwinBundle() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "darwin-bundle"
	createPackagingFormatDirectory(packagingFormat)
	bundleDirectoryPath := packagingFormatPath(packagingFormat)
	bundleContentsDirectoryPath, err := filepath.Abs(filepath.Join(bundleDirectoryPath, projectName+".app", "Contents"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for Contents directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(bundleContentsDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create Contents directory %s: %v", bundleContentsDirectoryPath, err)
		os.Exit(1)
	}
	bundleMacOSDirectoryPath, err := filepath.Abs(filepath.Join(bundleContentsDirectoryPath, "MacOS"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for MacOS directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(bundleMacOSDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create MacOS directory %s: %v", bundleMacOSDirectoryPath, err)
		os.Exit(1)
	}
	bundleResourcesDirectoryPath, err := filepath.Abs(filepath.Join(bundleContentsDirectoryPath, "Resources"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for Resources directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(bundleResourcesDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create Resources directory %s: %v", bundleResourcesDirectoryPath, err)
		os.Exit(1)
	}

	infoFilePath, err := filepath.Abs(filepath.Join(bundleContentsDirectoryPath, "Info.plist"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for Info.plist file %s: %v", infoFilePath, err)
		os.Exit(1)
	}

	infoFile, err := os.Create(infoFilePath)
	if err != nil {
		log.Errorf("Failed to create Info.plist file %s: %v", infoFilePath, err)
		os.Exit(1)
	}
	infoFileContent := []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`,
		`<plist version="1.0">`,
		`<dict>`,
		`	<key>CFBundleDevelopmentRegion</key>`,
		`	<string>English</string>`,
		`	<key>CFBundleExecutable</key>`,
		`	<string>` + projectName + `</string>`,
		`	<key>CFBundleGetInfoString</key>`,
		`	<string>` + projectName + `</string>`,
		`	<key>CFBundleIconFile</key>`,
		`	<string>icon.icns</string>`,
		`	<key>CFBundleIdentifier</key>`,
		`	<string></string>`,
		`	<key>CFBundleInfoDictionaryVersion</key>`,
		`	<string>6.0</string>`,
		`	<key>CFBundleLongVersionString</key>`,
		`	<string></string>`,
		`	<key>CFBundleName</key>`,
		`	<string></string>`,
		`	<key>CFBundlePackageType</key>`,
		`	<string>APPL</string>`,
		`	<key>CFBundleShortVersionString</key>`,
		`	<string></string>`,
		`	<key>CFBundleSignature</key>`,
		`	<string>????</string>`,
		`	<key>CFBundleVersion</key>`,
		`	<string></string>`,
		`	<key>CSResourcesFileMapped</key>`,
		`	<true/>`,
		`	<key>NSHumanReadableCopyright</key>`,
		`	<string></string>`,
		`</dict>`,
		`</plist>`,
	}

	for _, line := range infoFileContent {
		if _, err := infoFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write Info.plist file: %v", err)
			os.Exit(1)
		}
	}
	err = infoFile.Close()
	if err != nil {
		log.Errorf("Could not close Info.plist file: %v", err)
		os.Exit(1)
	}
	createDockerfile(packagingFormat)

	printInitFinished(packagingFormat)
}

func BuildDarwinBundle() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "darwin-bundle"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	log.Infof("Packaging bundle in %s", tmpPath)

	err := copy.Copy(build.OutputDirectoryPath("darwin"), filepath.Join(tmpPath, projectName+".app", "Contents", "MacOS"))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	outputFileName := projectName + ".app"
	outputFilePath := filepath.Join(build.OutputDirectoryPath("darwin-bundle"), outputFileName)
	runDockerPackaging(tmpPath, packagingFormat, []string{"png2icns", projectName + ".app/Contents/Resources/icon.icns", projectName + ".app/Contents/MacOS/assets/icon.png"})

	err = os.RemoveAll(outputFilePath)
	if err != nil {
		log.Errorf("Could not remove previous bundle directory: %v", err)
		os.Exit(1)
	}
	err = os.Rename(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move bundle directory: %v", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpPath)
	if err != nil {
		log.Errorf("Could not remove temporary build directory: %v", err)
		os.Exit(1)
	}
	printPackagingFinished(packagingFormat)
}
