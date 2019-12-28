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

// InitDarwinPkg initializes the a darwin pkg packaging format.
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

	fileutils.ExecuteTemplateFromAssetsBox("packaging/darwin-pkg/PackageInfo.tmpl.tmpl", filepath.Join(basePkgDirectoryPath, "PackageInfo.tmpl"), fileutils.AssetsBox, getTemplateData(projectName, ""))
	fileutils.ExecuteTemplateFromAssetsBox("packaging/darwin-pkg/Distribution.tmpl.tmpl", filepath.Join(pkgDirectoryPath, "flat", "Distribution.tmpl"), fileutils.AssetsBox, getTemplateData(projectName, ""))

	createDockerfile(packagingFormat, []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install cpio git make g++ wget libxml2-dev libssl1.0-dev zlib1g-dev -y",
		"WORKDIR /tmp",
		"RUN git clone https://github.com/hogliux/bomutils && cd bomutils && make > /dev/null && make install > /dev/null",
		"RUN wget https://storage.googleapis.com/google-code-archive-downloads/v2/code.google.com/xar/xar-1.5.2.tar.gz && tar -zxvf xar-1.5.2.tar.gz > /dev/null && cd xar-1.5.2 && ./configure > /dev/null && make > /dev/null && make install > /dev/null",
	})

	printInitFinished(packagingFormat)
}

// BuildDarwinPkg uses the InitDarwinPkg template to create an pkg package.
func BuildDarwinPkg(buildVersion string) {
	log.Infof("Building darwin-bundle first")
	BuildDarwinBundle(buildVersion)
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
	fileutils.CopyTemplateDir(packagingFormatPath(packagingFormat), filepath.Join(tmpPath), getTemplateData(projectName, buildVersion))
	outputFileName := projectName + "-" + buildVersion + ".pkg"
	runDockerPackaging(tmpPath, packagingFormat, []string{
		"(cd flat/root && find . | cpio -o --format odc --owner 0:80 | gzip -c ) > flat/base.pkg/Payload",
		"&&", "mkbom -u 0 -g 80 flat/root flat/base.pkg/Bom",
		"&&", "(cd flat && xar --compression none -cf '../" + outputFileName + "' * )",
	})

	outputFilePath := filepath.Join(build.OutputDirectoryPath("darwin-pkg"), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move pkg: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(packagingFormat)
}
