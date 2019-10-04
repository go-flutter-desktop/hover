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

	templateData := map[string]string{
		"projectName":      projectName,
		"organizationName": androidmanifest.AndroidOrganizationName(),
		"version":          pubspec.GetPubSpec().Version,
	}

	fileutils.CopyTemplate("packaging/PackageInfo.tmpl", filepath.Join(basePkgDirectoryPath, "PackageInfo"), fileutils.AssetsBox, templateData)
	fileutils.CopyTemplate("packaging/Distribution.tmpl", filepath.Join(pkgDirectoryPath, "flat", "Distribution"), fileutils.AssetsBox, templateData)

	createDockerfile(packagingFormat)

	printInitFinished(packagingFormat)
}

func BuildDarwinPkg() {
	log.Infof("Building darwin-bundle first")
	BuildDarwinBundle()
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "darwin-pkg"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
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
	runDockerPackaging(tmpPath, packagingFormat, []string{
		"(cd flat/root && find . | cpio -o --format odc --owner 0:80 | gzip -c ) > flat/base.pkg/Payload",
		"&&", "mkbom -u 0 -g 80 flat/root flat/base.pkg/Bom",
		"&&", "(cd flat && xar --compression none -cf '../" + outputFileName + "' * )",
	})

	outputFilePath := filepath.Join(build.OutputDirectoryPath("darwin-pkg"), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move pkg directory: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(packagingFormat)
}
