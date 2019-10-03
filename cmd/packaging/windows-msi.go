package packaging

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

var directoriesFileContent = []string{}
var directoryRefsFileContent = []string{}
var componentRefsFileContent = []string{}

func InitWindowsMsi() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "windows-msi"
	createPackagingFormatDirectory(packagingFormat)
	msiDirectoryPath := packagingFormatPath(packagingFormat)

	wxsFilePath, err := filepath.Abs(filepath.Join(msiDirectoryPath, projectName+".wxs"))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for %s.wxs file %s: %v", projectName, wxsFilePath, err)
		os.Exit(1)
	}
	wxsFile, err := os.Create(wxsFilePath)
	if err != nil {
		log.Errorf("Failed to create %s.wxs file %s: %v", projectName, wxsFilePath, err)
		os.Exit(1)
	}
	wxsFileContent := []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">`,
		fmt.Sprintf(`    <Product Id="*" UpgradeCode="*" Version="%s" Language="1033" Name="%s" Manufacturer="%s">`, pubspec.GetPubSpec().Version, projectName, getAuthor()),
		`        <Package InstallerVersion="300" Compressed="yes"/>`,
		fmt.Sprintf(`        <Media Id="1" Cabinet="%s.cab" EmbedCab="yes" />`, projectName),
		`        <Directory Id="TARGETDIR" Name="SourceDir">`,
		`            <Directory Id="ProgramFilesFolder">`,
		fmt.Sprintf(`                <Directory Id="APPLICATIONROOTDIRECTORY" Name="%s">`, projectName),
		`                    <Directory Id="ASSETSDIRECTORY" Name="assets"/>`,
		`                    <Directory Id="FLUTTERASSETSDIRECTORY" Name="flutter_assets">`,
		`                        <?include directories.wxi ?>`,
		`                    </Directory>`,
		`                </Directory>`,
		`            </Directory>`,
		`            <Directory Id="ProgramMenuFolder">`,
		fmt.Sprintf(`                <Directory Id="ApplicationProgramsFolder" Name="%s"/>`, projectName),
		`            </Directory>`,
		`        </Directory>`,
		`        <Icon Id="ShortcutIcon" SourceFile="build/assets/icon.ico"/>`,
		`        <DirectoryRef Id="APPLICATIONROOTDIRECTORY">`,
		fmt.Sprintf(`            <Component Id="%s.exe" Guid="*">`, projectName),
		fmt.Sprintf(`                <File Id="%s.exe" Source="build/%s.exe" KeyPath="yes"/>`, projectName, projectName),
		`            </Component>`,
		`            <Component Id="flutter_engine.dll" Guid="*">`,
		`                <File Id="flutter_engine.dll" Source="build/flutter_engine.dll" KeyPath="yes"/>`,
		`            </Component>`,
		`            <Component Id="icudtl.dat" Guid="*">`,
		`                <File Id="icudtl.dat" Source="build/icudtl.dat" KeyPath="yes"/>`,
		`            </Component>`,
		`        </DirectoryRef>`,
		`        <DirectoryRef Id="ASSETSDIRECTORY">`,
		`            <Component Id="icon.png" Guid="*">`,
		`                <File Id="icon.png" Source="build/assets/icon.png" KeyPath="yes"/>`,
		`            </Component>`,
		`        </DirectoryRef>`,
		`        <?include directory_refs.wxi ?>`,
		`        <DirectoryRef Id="ApplicationProgramsFolder">`,
		`            <Component Id="ApplicationShortcut" Guid="*">`,
		`                <Shortcut Id="ApplicationStartMenuShortcut"`,
		fmt.Sprintf(`                        Name="%s"`, projectName),
		fmt.Sprintf(`                        Description="%s"`, pubspec.GetPubSpec().Description),
		fmt.Sprintf(`                        Target="[#%s.exe]"`, projectName),
		`                        WorkingDirectory="APPLICATIONROOTDIRECTORY"`,
		`                        Icon="ShortcutIcon"/>`,
		`                <RemoveFolder Id="CleanUpShortCut" On="uninstall"/>`,
		fmt.Sprintf(`                <RegistryValue Root="HKCU" Key="Software\%s\%s" Name="installed" Type="integer" Value="1" KeyPath="yes"/>`, getAuthor(), projectName),
		`            </Component>`,
		`        </DirectoryRef>`,
		fmt.Sprintf(`        <Feature Id="MainApplication" Title="%s" Level="1">`, projectName),
		fmt.Sprintf(`            <ComponentRef Id="%s.exe"/>`, projectName),
		`            <ComponentRef Id="flutter_engine.dll"/>`,
		`            <ComponentRef Id="icudtl.dat"/>`,
		`            <ComponentRef Id="icon.png"/>`,
		`            <ComponentRef Id="ApplicationShortcut"/>`,
		`            <?include component_refs.wxi ?>`,
		`        </Feature>`,
		`    </Product>`,
		`</Wix>`,
	}

	for _, line := range wxsFileContent {
		if _, err := wxsFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write %s.wxs: %v", projectName, err)
			os.Exit(1)
		}
	}
	err = wxsFile.Close()
	if err != nil {
		log.Errorf("Could not close %s.wxs: %v", projectName, err)
		os.Exit(1)
	}

	createDockerfile(packagingFormat)

	printInitFinished(packagingFormat)
}

func BuildWindowsMsi() {
	projectName := pubspec.GetPubSpec().Name
	packagingFormat := "windows-msi"
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
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
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		log.Errorf("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}
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

	outputFileName := projectName + ".msi"
	outputFilePath := filepath.Join(build.OutputDirectoryPath("windows-msi"), outputFileName)
	runDockerPackaging(tmpPath, packagingFormat, []string{"convert", "-resize", "x16", "build/assets/icon.png", "build/assets/icon.ico", "&&", "wixl", "-v", projectName + ".wxs"})

	err = copy.Copy(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Errorf("Could not move msi file: %v", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpPath)
	if err != nil {
		log.Errorf("Could not remove temporary build directory: %v", err)
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
