package packaging

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

// InitLinuxRpm initializes the linux rpm packaging format.
func InitLinuxRpm() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-rpm"
	createPackagingFormatDirectory(packagingFormat)
	rpmDirectoryPath := packagingFormatPath(packagingFormat)

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
	rpmRootSourcesDirectoryPath, err := filepath.Abs(filepath.Join(rpmBuildRootDirectoryPath, "{{.strippedProjectName}}-{{.version}}-{{.version}}.x86_64"))
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
	binFilePath := filepath.Join(binDirectoryPath, removeDashesAndUnderscores(projectName))

	fileutils.ExecuteTemplateFromAssetsBox("packaging/linux-rpm/app.spec.tmpl.tmpl", filepath.Join(rpmSpecsDirectoryPath, removeDashesAndUnderscores(projectName)+".spec.tmpl"), fileutils.AssetsBox, getTemplateData(projectName, ""))
	fileutils.ExecuteTemplateFromAssetsBox("packaging/linux/bin.tmpl", binFilePath, fileutils.AssetsBox, getTemplateData(projectName, ""))

	err = os.Chmod(binFilePath, 0777)
	if err != nil {
		log.Errorf("Failed to change file permissions for bin file: %v", err)
		os.Exit(1)
	}

	createLinuxDesktopFile(filepath.Join(applicationsDirectoryPath, projectName+".desktop"), "/usr/bin/"+removeDashesAndUnderscores(projectName), "/usr/lib/"+projectName+"/assets/icon.png")

	createDockerfile(packagingFormat, []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install rpm file -y",
	})

	printInitFinished(packagingFormat)
}

// BuildLinuxRpm uses the InitLinuxRpm template to create a rpm package.
func BuildLinuxRpm(buildVersion string) {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-rpm"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
	log.Infof("Packaging rpm in %s", tmpPath)

	libDirectoryPath, err := filepath.Abs(filepath.Join(tmpPath, "rpmbuild", "BUILDROOT", fmt.Sprintf("%s-%s-%s.x86_64", removeDashesAndUnderscores(projectName), buildVersion, buildVersion), "usr", "lib"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for lib directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(libDirectoryPath, 0755)
	if err != nil {
		log.Errorf("Cannot create the lib directory: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(build.OutputDirectoryPath("linux"), filepath.Join(libDirectoryPath, projectName))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	fileutils.CopyTemplateDir(packagingFormatPath(packagingFormat), filepath.Join(tmpPath), getTemplateData(projectName, buildVersion))
	resultFileName := fmt.Sprintf("%s-%s-%s.x86_64.rpm", removeDashesAndUnderscores(projectName), buildVersion, buildVersion)
	outputFileName := projectName + "-" + buildVersion + ".rpm"
	runDockerPackaging(tmpPath, packagingFormat, []string{"rpmbuild --define '_topdir /app/rpmbuild' -ba /app/rpmbuild/SPECS/" + removeDashesAndUnderscores(projectName) + ".spec", "&&", "rm /root/.rpmdb -r"})

	outputFilePath := filepath.Join(build.OutputDirectoryPath("linux-rpm"), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, "rpmbuild", "RPMS", "x86_64", resultFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move rpm file: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(packagingFormat)
}
