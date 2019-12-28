package packaging

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

var directoriesFileContent []string
var directoryRefsFileContent []string
var componentRefsFileContent []string

// InitWindowsMsi initializes the windows msi packaging format.
func InitWindowsMsi() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "windows-msi"
	createPackagingFormatDirectory(packagingFormat)
	msiDirectoryPath := packagingFormatPath(packagingFormat)

	fileutils.ExecuteTemplateFromAssetsBox("packaging/windows-msi/app.wxs.tmpl.tmpl", filepath.Join(msiDirectoryPath, projectName+".wxs.tmpl"), fileutils.AssetsBox, getTemplateData(projectName, ""))

	createDockerfile(packagingFormat, []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install wixl imagemagick -y",
	})

	printInitFinished(packagingFormat)
}

// BuildWindowsMsi uses the InitWindowsMsi template to create a msi package.
func BuildWindowsMsi(buildVersion string) {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "windows-msi"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
	log.Infof("Packaging msi in %s", tmpPath)

	buildDirectoryPath, err := filepath.Abs(filepath.Join(tmpPath, "build"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for build directory: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(build.OutputDirectoryPath("windows"), filepath.Join(buildDirectoryPath))
	if err != nil {
		log.Errorf("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	fileutils.CopyTemplateDir(packagingFormatPath(packagingFormat), filepath.Join(tmpPath), getTemplateData(projectName, buildVersion))
	directoriesFilePath, err := filepath.Abs(filepath.Join(tmpPath, "directories.wxi"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for directories.wxi file %s: %v", projectName, err)
		os.Exit(1)
	}
	directoriesFile, err := os.Create(directoriesFilePath)
	if err != nil {
		log.Errorf("Failed to create directories.wxi file %s: %v", projectName, err)
		os.Exit(1)
	}
	directoryRefsFilePath, err := filepath.Abs(filepath.Join(tmpPath, "directory_refs.wxi"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for directory_refs.wxi file %s: %v", projectName, err)
		os.Exit(1)
	}
	directoryRefsFile, err := os.Create(directoryRefsFilePath)
	if err != nil {
		log.Errorf("Failed to create directory_refs.wxi file %s: %v", projectName, err)
		os.Exit(1)
	}
	componentRefsFilePath, err := filepath.Abs(filepath.Join(tmpPath, "component_refs.wxi"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for component_refs.wxi file %s: %v", projectName, err)
		os.Exit(1)
	}
	componentRefsFile, err := os.Create(componentRefsFilePath)
	if err != nil {
		log.Errorf("Failed to create component_refs.wxi file %s: %v", projectName, err)
		os.Exit(1)
	}
	directoriesFileContent = append(directoriesFileContent, `<Include>`)
	directoryRefsFileContent = append(directoryRefsFileContent, `<Include>`)
	componentRefsFileContent = append(componentRefsFileContent, `<Include>`)
	processFiles(filepath.Join(buildDirectoryPath, "flutter_assets"))
	directoriesFileContent = append(directoriesFileContent, `</Include>`)
	directoryRefsFileContent = append(directoryRefsFileContent, `</Include>`)
	componentRefsFileContent = append(componentRefsFileContent, `</Include>`)

	for _, line := range directoriesFileContent {
		if _, err := directoriesFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write directories.wxi: %v", projectName, err)
			os.Exit(1)
		}
	}
	err = directoriesFile.Close()
	if err != nil {
		log.Errorf("Could not close directories.wxi: %v", projectName, err)
		os.Exit(1)
	}
	for _, line := range directoryRefsFileContent {
		if _, err := directoryRefsFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write directory_refs.wxi: %v", projectName, err)
			os.Exit(1)
		}
	}
	err = directoryRefsFile.Close()
	if err != nil {
		log.Errorf("Could not close directory_refs.wxi: %v", projectName, err)
		os.Exit(1)
	}
	for _, line := range componentRefsFileContent {
		if _, err := componentRefsFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write component_refs.wxi: %v", projectName, err)
			os.Exit(1)
		}
	}
	err = componentRefsFile.Close()
	if err != nil {
		log.Errorf("Could not close component_refs.wxi: %v", projectName, err)
		os.Exit(1)
	}

	runDockerPackaging(tmpPath, packagingFormat, []string{"convert", "-resize", "x16", "build/assets/icon.png", "build/assets/icon.ico", "&&", "wixl", "-v", projectName + ".wxs"})

	resultFileName := projectName + ".msi"
	outputFileName := projectName + "-" + buildVersion + ".msi"
	outputFilePath := filepath.Join(build.OutputDirectoryPath("windows-msi"), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, resultFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move msi file: %v", err)
		os.Exit(1)
	}

	printPackagingFinished(packagingFormat)
}

func processFiles(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Errorf("Failed to read directory %s: %v", path, err)
		os.Exit(1)
	}

	for _, f := range files {
		p := filepath.Join(path, f.Name())
		relativePath := strings.Split(strings.Split(p, "flutter_assets"+string(filepath.Separator))[1], string(filepath.Separator))
		if f.IsDir() {
			directoriesFileContent = append(directoriesFileContent,
				`<Directory Id="FLUTTERASSETSDIRECTORY_`+strings.Join(relativePath, "_")+`" Name="`+f.Name()+`">`,
			)
			processFiles(p)
			directoriesFileContent = append(directoriesFileContent,
				`</Directory>`,
			)
		} else {
			if len(relativePath) > 1 {
				directoryRefsFileContent = append(directoryRefsFileContent,
					`<DirectoryRef Id="FLUTTERASSETSDIRECTORY_`+strings.Join(relativePath[:len(relativePath)-1], "_")+`">`,
				)
			} else {
				directoryRefsFileContent = append(directoryRefsFileContent,
					`<DirectoryRef Id="FLUTTERASSETSDIRECTORY">`,
				)
			}
			directoryRefsFileContent = append(directoryRefsFileContent,
				`<Component Id="flutter_assets_`+strings.Join(relativePath, "_")+`" Guid="*">`,
				`<File Id="flutter_assets_`+strings.Join(relativePath, "_")+`" Source="build/flutter_assets/`+strings.Join(relativePath, "/")+`" KeyPath="yes"/>`,
				`</Component>`,
				`</DirectoryRef>`,
			)
			componentRefsFileContent = append(componentRefsFileContent,
				`<ComponentRef Id="flutter_assets_`+strings.Join(relativePath, "_")+`"/>`,
			)
		}
	}
}
