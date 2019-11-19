package packaging

import (
	"github.com/otiai10/copy"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

// InitLinuxDeb initialize the a linux deb packagingFormat.
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

	binFilePath := filepath.Join(binDirectoryPath, removeDashesAndUnderscores(projectName))

	fileutils.CopyTemplateFromAssetsBox("packaging/control.tmpl", filepath.Join(debDebianDirectoryPath, "control.tmpl"), fileutils.AssetsBox, getTemplateData(projectName))
	fileutils.CopyTemplateFromAssetsBox("packaging/bin.tmpl", binFilePath, fileutils.AssetsBox, getTemplateData(projectName))

	err = os.Chmod(binFilePath, 0777)
	if err != nil {
		log.Errorf("Failed to change file permissions for bin file: %v", err)
		os.Exit(1)
	}

	createLinuxDesktopFile(filepath.Join(applicationsDirectoryPath, projectName+".desktop"), "/usr/bin/"+removeDashesAndUnderscores(projectName), "/usr/lib/"+projectName+"/assets/icon.png")
	createDockerfile(packagingFormat, []string{
		"FROM ubuntu:bionic",
	})

	printInitFinished(packagingFormat)
}

// BuildLinuxDeb uses the InitLinuxDeb template to create a deb package.
func BuildLinuxDeb() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-deb"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
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
	fileutils.CopyTemplateDir(packagingFormatPath(packagingFormat), filepath.Join(tmpPath), getTemplateData(projectName))

	outputFileName := removeDashesAndUnderscores(projectName) + "_" + runtime.GOARCH + ".deb"
	runDockerPackaging(tmpPath, packagingFormat, []string{"dpkg-deb", "--build", ".", outputFileName})

	outputFilePath := filepath.Join(build.OutputDirectoryPath("linux-deb"), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move deb file: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(packagingFormat)
}
