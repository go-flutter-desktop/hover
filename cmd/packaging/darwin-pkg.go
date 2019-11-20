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

func InitDarwinPkg(buildTarget build.Target) {
	projectName := pubspec.GetPubSpec().Name

	createPackagingFormatDirectory(buildTarget)
	pkgDirectoryPath := PackagingFormatPath(buildTarget)

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

	createDockerfile(packagingFormat, []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install cpio git make g++ wget libxml2-dev libssl1.0-dev zlib1g-dev -y",
		"WORKDIR /tmp",
		"RUN git clone https://github.com/hogliux/bomutils && cd bomutils && make > /dev/null && make install > /dev/null",
		"RUN wget https://storage.googleapis.com/google-code-archive-downloads/v2/code.google.com/xar/xar-1.5.2.tar.gz && tar -zxvf xar-1.5.2.tar.gz > /dev/null && cd xar-1.5.2 && ./configure > /dev/null && make > /dev/null && make install > /dev/null",
	})

	printInitFinished(buildTarget)
}

func BuildDarwinPkg(buildTarget build.Target) {
	log.Infof("Building darwin-bundle first")
	projectName := pubspec.GetPubSpec().Name
	tmpPath := getTemporaryBuildDirectory(projectName, buildTarget)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
	log.Infof("Packaging pkg in %s", tmpPath)

	err := copy.Copy(build.OutputDirectoryPath(build.Target{
		Platform:        build.TargetPlatforms.Darwin,
		PackagingFormat: build.TargetPackagingFormats.Bundle,
	}, false), filepath.Join(tmpPath, "flat", "root", "Applications"))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(PackagingFormatPath(buildTarget), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	outputFileName := projectName + " " + pubspec.GetPubSpec().Version + " Installer.pkg"
	runDockerPackaging(tmpPath, buildTarget, []string{
		"(cd flat/root && find . | cpio -o --format odc --owner 0:80 | gzip -c ) > flat/base.pkg/Payload",
		"&&", "mkbom -u 0 -g 80 flat/root flat/base.pkg/Bom",
		"&&", "(cd flat && xar --compression none -cf '../" + outputFileName + "' * )",
	})

	outputFilePath := filepath.Join(build.OutputDirectoryPath(buildTarget, true), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move pkg: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(buildTarget)
}
