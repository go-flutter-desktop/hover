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

// InitLinuxSnap initialize the a linux snap packaging format
func InitLinuxSnap(buildTarget build.Target) {
	projectName := pubspec.GetPubSpec().Name
	createPackagingFormatDirectory(buildTarget)
	snapDirectoryPath := PackagingFormatPath(buildTarget)

	snapLocalDirectoryPath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "snap", "local"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for snap local directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(snapLocalDirectoryPath, 0775)
	if err != nil {
		log.Errorf("Failed to create snap local directory %s: %v", snapDirectoryPath, err)
		os.Exit(1)
	}

	templateData := map[string]string{
		"projectName":         projectName,
		"strippedProjectName": strings.ToLower(removeDashesAndUnderscores(projectName)),
		"version":             pubspec.GetPubSpec().Version,
		"description":         pubspec.GetPubSpec().Description,
		"dependencies":        strings.Join(linuxPackagingDependencies, "\n      - "),
	}

	fileutils.CopyTemplate("packaging/snapcraft.yaml.tmpl", filepath.Join(snapDirectoryPath, "snap", "snapcraft.yaml"), fileutils.AssetsBox, templateData)

	createLinuxDesktopFile(filepath.Join(snapLocalDirectoryPath, projectName+".desktop"), "/"+projectName, "/icon.png")
	createDockerfile(packagingFormat, []string{
		"FROM snapcore/snapcraft",
	})

	printInitFinished(buildTarget)
}

// BuildLinuxSnap uses the InitLinuxSnap template to create a snap package.
func BuildLinuxSnap(buildTarget build.Target) {
	projectName := pubspec.GetPubSpec().Name
	tmpPath := getTemporaryBuildDirectory(projectName, buildTarget)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
	log.Infof("Packaging snap in %s", tmpPath)

	err := copy.Copy(filepath.Join(build.BuildPath, "assets"), filepath.Join(tmpPath, "assets"))
	if err != nil {
		log.Errorf("Could not copy assets folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(build.OutputDirectoryPath(buildTarget, false), filepath.Join(tmpPath, "build"))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(PackagingFormatPath(buildTarget), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	runDockerPackaging(tmpPath, buildTarget, []string{"snapcraft"})

	outputFileName := strings.ToLower(removeDashesAndUnderscores(projectName)) + "_" + pubspec.GetPubSpec().Version + "_" + runtime.GOARCH + ".snap"
	outputFilePath := filepath.Join(build.OutputDirectoryPath(buildTarget, true), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move snap file: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(buildTarget)
}
