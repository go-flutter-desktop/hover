package packaging

import (
	"os"
	"path/filepath"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

func InitLinuxAppImage() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-appimage"
	createPackagingFormatDirectory(packagingFormat)
	appImageDirectoryPath := packagingFormatPath(packagingFormat)

	templateData := map[string]string{
		"projectName": projectName,
	}

	appRunFilePath := filepath.Join(appImageDirectoryPath, "AppRun")

	fileutils.CopyTemplate("packaging/AppRun.tmpl", filepath.Join(appImageDirectoryPath, "AppRun"), fileutils.AssetsBox, templateData)

	err := os.Chmod(appRunFilePath, 0777)
	if err != nil {
		log.Errorf("Failed to change file permissions for AppRun file: %v", err)
		os.Exit(1)
	}

	createLinuxDesktopFile(filepath.Join(appImageDirectoryPath, projectName+".desktop"), "", "/build/assets/icon")
	createDockerfile(packagingFormat)

	printInitFinished(packagingFormat)
}

func BuildLinuxAppImage() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "linux-appimage"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
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

	runDockerPackaging(tmpPath, packagingFormat, []string{"appimagetool", "."})

	outputFileName := projectName + "-x86_64.AppImage"
	outputFilePath := filepath.Join(build.OutputDirectoryPath("linux-appimage"), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move AppImage file: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(packagingFormat)
}
