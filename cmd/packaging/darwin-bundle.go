package packaging

import (
	"os"
	"path/filepath"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/androidmanifest"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

func InitDarwinBundle(buildTarget build.Target) {
	projectName := pubspec.GetPubSpec().Name
	createPackagingFormatDirectory(buildTarget)
	bundleDirectoryPath := PackagingFormatPath(buildTarget)
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

	templateData := map[string]string{
		"projectName":      projectName,
		"organizationName": androidmanifest.AndroidOrganizationName(),
		"version":          pubspec.GetPubSpec().Version,
		"description":      pubspec.GetPubSpec().Description,
	}

	fileutils.CopyTemplate("packaging/Info.plist.tmpl", filepath.Join(bundleContentsDirectoryPath, "Info.plist"), fileutils.AssetsBox, templateData)

	createDockerfile(buildTarget)

	printInitFinished(buildTarget)
}

func BuildDarwinBundle(buildTarget build.Target) {
	projectName := pubspec.GetPubSpec().Name
	tmpPath := getTemporaryBuildDirectory(projectName, buildTarget)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
	log.Infof("Packaging bundle in %s", tmpPath)

	err := copy.Copy(build.OutputDirectoryPath(buildTarget, false), filepath.Join(tmpPath, projectName+".app", "Contents", "MacOS"))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(PackagingFormatPath(buildTarget), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	runDockerPackaging(tmpPath, buildTarget, []string{"mkdir", "-p", projectName + ".app/Contents/Resources", "&&", "png2icns", projectName + ".app/Contents/Resources/icon.icns", projectName + ".app/Contents/MacOS/assets/icon.png"})

	outputFileName := projectName + ".app"
	outputFilePath := filepath.Join(build.OutputDirectoryPath(buildTarget, true), outputFileName)
	err = os.RemoveAll(outputFilePath)
	if err != nil {
		log.Errorf("Could not remove previous bundle directory: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move bundle directory: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(buildTarget)
}
