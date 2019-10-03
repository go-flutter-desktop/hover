package packaging

import (
	"os"
	"path/filepath"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/androidmanifest"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

func InitDarwinPkg() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "darwin-pkg"
	if _, err := os.Stat(packagingFormatPath("darwin-bundle")); os.IsNotExist(err) {
		log.Errorf("You have to init `darwin-bundle` first. Run `hover init-packaging darwin-bundle`")
		os.Exit(1)
	}
	createPackagingFormatDirectory(packagingFormat)
	pkgDirectoryPath := packagingFormatPath(packagingFormat)

	basePkgDirectoryPath, err := filepath.Abs(filepath.Join(pkgDirectoryPath, "flat", "base.pkg"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for base.pkg directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(basePkgDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create base.pkg directory %s: %v", basePkgDirectoryPath, err)
		os.Exit(1)
	}

	packageInfoFilePath, err := filepath.Abs(filepath.Join(basePkgDirectoryPath, "PackageInfo"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for PackageInfo file %s: %v", packageInfoFilePath, err)
		os.Exit(1)
	}
	packageInfoFile, err := os.Create(packageInfoFilePath)
	if err != nil {
		log.Errorf("Failed to create PackageInfo file %s: %v", packageInfoFilePath, err)
		os.Exit(1)
	}
	packageInfoFileContent := []string{
		`<pkg-info format-version="2" identifier="` + androidmanifest.AndroidOrganizationName() + `.base.pkg" version="` + pubspec.GetPubSpec().Version + `" install-location="/" auth="root">`,
		`	<bundle-version>`,
		`		<bundle id="` + androidmanifest.AndroidOrganizationName() + `" CFBundleIdentifier="` + androidmanifest.AndroidOrganizationName() + `" path="./Applications/` + projectName + `.app" CFBundleVersion="` + pubspec.GetPubSpec().Version + `"/>`,
		`	</bundle-version>`,
		`</pkg-info>`,
	}

	for _, line := range packageInfoFileContent {
		if _, err := packageInfoFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write PackageInfo file: %v", err)
			os.Exit(1)
		}
	}
	err = packageInfoFile.Close()
	if err != nil {
		log.Errorf("Could not close PackageInfo file: %v", err)
		os.Exit(1)
	}

	distributionFilePath, err := filepath.Abs(filepath.Join(pkgDirectoryPath, "flat", "Distribution"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for Distribution file %s: %v", distributionFilePath, err)
		os.Exit(1)
	}
	distributionFile, err := os.Create(distributionFilePath)
	if err != nil {
		log.Errorf("Failed to create Distribution file %s: %v", packageInfoFilePath, err)
		os.Exit(1)
	}
	distributionFileContent := []string{
		`<?xml version="1.0" encoding="utf-8"?>`,
		`<installer-gui-script minSpecVersion="1">`,
		`	<title>` + projectName + `</title>`,
		`	<background alignment="topleft" file="root/Applications/` + projectName + `.app/Contents/MacOS/assets/icon.png"/>`,
		`	<choices-outline>`,
		`		<line choice="choiceBase"/>`,
		`	</choices-outline>`,
		`	<choice id="choiceBase" title="base">`,
		`		<pkg-ref id="` + androidmanifest.AndroidOrganizationName() + `.base.pkg"/>`,
		`	</choice>`,
		`	<pkg-ref id="` + androidmanifest.AndroidOrganizationName() + `.base.pkg" version="` + pubspec.GetPubSpec().Version + `" auth="Root">#base.pkg</pkg-ref>`,
		`</installer-gui-script>`,
	}

	for _, line := range distributionFileContent {
		if _, err := distributionFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write Distribution file: %v", err)
			os.Exit(1)
		}
	}
	err = distributionFile.Close()
	if err != nil {
		log.Errorf("Could not close Distribution file: %v", err)
		os.Exit(1)
	}

	createDockerfile(packagingFormat)

	printInitFinished(packagingFormat)
}

func BuildDarwinPkg() {
	log.Infof("Building darwin-bundle first")
	BuildDarwinBundle()
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "darwin-pkg"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	log.Infof("Packaging pkg in %s", tmpPath)

	err := copy.Copy(build.OutputDirectoryPath("darwin-bundle"), filepath.Join(tmpPath, "flat", "root", "Applications"))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	outputFileName := projectName + " " + pubspec.GetPubSpec().Version + " Installer.pkg"
	outputFilePath := filepath.Join(build.OutputDirectoryPath("darwin-pkg"), outputFileName)
	runDockerPackaging(tmpPath, packagingFormat, []string{
		"(cd flat/root && find . | cpio -o --format odc --owner 0:80 | gzip -c ) > flat/base.pkg/Payload",
		"&&", "mkbom -u 0 -g 80 flat/root flat/base.pkg/Bom",
		"&&", "(cd flat && xar --compression none -cf '../" + projectName + " " + pubspec.GetPubSpec().Version + " Installer.pkg' * )",
	})
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move pkg directory: %v", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpPath)
	if err != nil {
		log.Errorf("Could not remove temporary build directory: %v", err)
		os.Exit(1)
	}
	printPackagingFinished(packagingFormat)
}
