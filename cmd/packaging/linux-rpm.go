package packaging

import (
	"fmt"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
	"github.com/otiai10/copy"
	"os"
	"path/filepath"
)

// InitLinuxRpm initialize the a linux rpm packaging format.
func InitLinuxRpm(buildTarget build.Target) {
	projectName := pubspec.GetPubSpec().Name
	createPackagingFormatDirectory(buildTarget)
	rpmDirectoryPath := PackagingFormatPath(buildTarget)

	rpmRmpbuildDirectoryPath, err := filepath.Abs(filepath.Join(rpmDirectoryPath, "rpmbuild"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for rpmbuild directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(rpmRmpbuildDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create rpmbuild directory %s: %v", rpmRmpbuildDirectoryPath, err)
		os.Exit(1)
	}
	rpmSpecsDirectoryPath, err := filepath.Abs(filepath.Join(rpmRmpbuildDirectoryPath, "SPECS"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for SPECS directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(rpmSpecsDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create SPECS directory %s: %v", rpmSpecsDirectoryPath, err)
		os.Exit(1)
	}
	rpmBuildRootDirectoryPath, err := filepath.Abs(filepath.Join(rpmRmpbuildDirectoryPath, "BUILDROOT"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for BUILDROOT directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(rpmBuildRootDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create BUILDROOT directory %s: %v", rpmBuildRootDirectoryPath, err)
		os.Exit(1)
	}
	rpmRootSourcesDirectoryPath, err := filepath.Abs(filepath.Join(rpmBuildRootDirectoryPath, fmt.Sprintf("%s-%s-%s.x86_64", removeDashesAndUnderscores(projectName), pubspec.GetPubSpec().Version, pubspec.GetPubSpec().Version)))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for BUILDROOT directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(rpmRootSourcesDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create root sources directory %s: %v", rpmRootSourcesDirectoryPath, err)
		os.Exit(1)
	}

	binDirectoryPath, err := filepath.Abs(filepath.Join(rpmRootSourcesDirectoryPath, "usr", "bin"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for bin directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(binDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create bin directory %s: %v", binDirectoryPath, err)
		os.Exit(1)
	}
	applicationsDirectoryPath, err := filepath.Abs(filepath.Join(rpmRootSourcesDirectoryPath, "usr", "share", "applications"))
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
		"version":             pubspec.GetPubSpec().Version,
		"description":         pubspec.GetPubSpec().Description,
	}

	binFilePath := filepath.Join(binDirectoryPath, removeDashesAndUnderscores(projectName))

	fileutils.CopyTemplate("packaging/app.spec.tmpl", filepath.Join(rpmSpecsDirectoryPath, removeDashesAndUnderscores(projectName)+".spec"), fileutils.AssetsBox, templateData)
	fileutils.CopyTemplate("packaging/bin.tmpl", binFilePath, fileutils.AssetsBox, templateData)

	err = os.Chmod(binFilePath, 0777)
	if err != nil {
		log.Errorf("Failed to change file permissions for bin file: %v", err)
		os.Exit(1)
	}

	createLinuxDesktopFile(filepath.Join(applicationsDirectoryPath, projectName+".desktop"), "/usr/bin/"+removeDashesAndUnderscores(projectName), "/usr/lib/"+projectName+"/assets/icon.png")

	createDockerfile(buildTarget, []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install rpm -y",
	})

	printInitFinished(buildTarget)
}

// BuildLinuxRpm uses the InitLinuxRpm template to create a rpm package.
func BuildLinuxRpm(buildTarget build.Target) {
	projectName := pubspec.GetPubSpec().Name
	tmpPath := getTemporaryBuildDirectory(projectName, buildTarget)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
	log.Infof("Packaging rpm in %s", tmpPath)

	libDirectoryPath, err := filepath.Abs(filepath.Join(tmpPath, "rpmbuild", "BUILDROOT", fmt.Sprintf("%s-%s-%s.x86_64", removeDashesAndUnderscores(projectName), pubspec.GetPubSpec().Version, pubspec.GetPubSpec().Version), "usr", "lib"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for lib directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(libDirectoryPath, 0755)
	if err != nil {
		log.Errorf("Cannot create the lib directory: %v", err)
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

	outputFileName := fmt.Sprintf("%s-%s-%s.x86_64.rpm", removeDashesAndUnderscores(projectName), pubspec.GetPubSpec().Version, pubspec.GetPubSpec().Version)
	runDockerPackaging(tmpPath, buildTarget, []string{"rpmbuild --define '_topdir /app/rpmbuild' -ba /app/rpmbuild/SPECS/" + removeDashesAndUnderscores(projectName) + ".spec", "&&", "rm /root/.rpmdb -r"})

	outputFilePath := filepath.Join(build.OutputDirectoryPath(buildTarget, true), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, "rpmbuild", "RPMS", "x86_64", outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move rpm file: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(buildTarget)
}
