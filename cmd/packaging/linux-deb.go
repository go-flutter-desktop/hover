package packaging

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

// InitLinuxDeb initialize the a linux deb packaging format.
func InitLinuxDeb(buildTarget build.Target) {
	projectName := pubspec.GetPubSpec().Name
	createPackagingFormatDirectory(buildTarget)
	debDirectoryPath := PackagingFormatPath(buildTarget)
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

	templateData := map[string]string{
		"projectName":         projectName,
		"strippedProjectName": removeDashesAndUnderscores(projectName),
		"author":              getAuthor(),
		"version":             pubspec.GetPubSpec().Version,
		"description":         pubspec.GetPubSpec().Description,
		"dependencies":        strings.Join(linuxPackagingDependencies, ","),
	}

	binFilePath := filepath.Join(binDirectoryPath, removeDashesAndUnderscores(projectName))

	fileutils.CopyTemplate("packaging/control.tmpl", filepath.Join(debDebianDirectoryPath, "control"), fileutils.AssetsBox, templateData)
	fileutils.CopyTemplate("packaging/bin.tmpl", binFilePath, fileutils.AssetsBox, templateData)

	err = os.Chmod(binFilePath, 0777)
	if err != nil {
		log.Errorf("Failed to change file permissions for bin file: %v", err)
		os.Exit(1)
	}

	createLinuxDesktopFile(filepath.Join(applicationsDirectoryPath, projectName+".desktop"), "/usr/bin/"+projectName, "/usr/lib/"+projectName+"/assets/icon.png")
	createDockerfile(buildTarget)

	printInitFinished(buildTarget)
}

// BuildLinuxDeb uses the InitLinuxDeb template to create a deb package.
func BuildLinuxDeb(buildTarget build.Target) {
	projectName := pubspec.GetPubSpec().Name
	tmpPath := getTemporaryBuildDirectory(projectName, buildTarget)
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
	err = copy.Copy(build.OutputDirectoryPath(buildTarget, false), filepath.Join(libDirectoryPath, projectName))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(PackagingFormatPath(buildTarget), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	outputFileName := removeDashesAndUnderscores(projectName) + "_" + runtime.GOARCH + ".deb"
	runDockerPackaging(tmpPath, buildTarget, []string{"dpkg-deb", "--build", ".", outputFileName})

	outputFilePath := filepath.Join(build.OutputDirectoryPath(buildTarget, true), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move deb file: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(buildTarget)
}
