package packaging

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-flutter-desktop/hover/internal/log"
)

var directoriesFileContent []string
var directoryRefsFileContent []string
var componentRefsFileContent []string

// WindowsMsiTask packaging for windows as msi
var WindowsMsiTask = &packagingTask{
	packagingFormatName: "windows-msi",
	templateFiles: map[string]string{
		"windows-msi/app.wxs.tmpl": "{{.packageName}}.wxs.tmpl",
	},
	buildOutputDirectory:          "build",
	packagingScriptTemplate:       "convert -resize x16 build/assets/icon.png build/assets/icon.ico && wixl -v {{.packageName}}.wxs && mv -p {{.packageName}}.msi \"{{.applicationName}}.msi\"",
	outputFileExtension:           "msi",
	outputFileContainsVersion:     false,
	outputFileUsesApplicationName: true,
	generateBuildFiles: func(packageName, tmpPath string) {
		directoriesFilePath, err := filepath.Abs(filepath.Join(tmpPath, "directories.wxi"))
		if err != nil {
			log.Errorf("Failed to resolve absolute path for directories.wxi file %s: %v", packageName, err)
			os.Exit(1)
		}
		directoriesFile, err := os.Create(directoriesFilePath)
		if err != nil {
			log.Errorf("Failed to create directories.wxi file %s: %v", packageName, err)
			os.Exit(1)
		}
		directoryRefsFilePath, err := filepath.Abs(filepath.Join(tmpPath, "directory_refs.wxi"))
		if err != nil {
			log.Errorf("Failed to resolve absolute path for directory_refs.wxi file %s: %v", packageName, err)
			os.Exit(1)
		}
		directoryRefsFile, err := os.Create(directoryRefsFilePath)
		if err != nil {
			log.Errorf("Failed to create directory_refs.wxi file %s: %v", packageName, err)
			os.Exit(1)
		}
		componentRefsFilePath, err := filepath.Abs(filepath.Join(tmpPath, "component_refs.wxi"))
		if err != nil {
			log.Errorf("Failed to resolve absolute path for component_refs.wxi file %s: %v", packageName, err)
			os.Exit(1)
		}
		componentRefsFile, err := os.Create(componentRefsFilePath)
		if err != nil {
			log.Errorf("Failed to create component_refs.wxi file %s: %v", packageName, err)
			os.Exit(1)
		}
		directoriesFileContent = append(directoriesFileContent, `<Include>`)
		directoryRefsFileContent = append(directoryRefsFileContent, `<Include>`)
		componentRefsFileContent = append(componentRefsFileContent, `<Include>`)
		windowsMsiProcessFiles(filepath.Join(tmpPath, "build", "flutter_assets"))
		directoriesFileContent = append(directoriesFileContent, `</Include>`)
		directoryRefsFileContent = append(directoryRefsFileContent, `</Include>`)
		componentRefsFileContent = append(componentRefsFileContent, `</Include>`)

		for _, line := range directoriesFileContent {
			if _, err := directoriesFile.WriteString(line + "\n"); err != nil {
				log.Errorf("Could not write directories.wxi: %v", packageName, err)
				os.Exit(1)
			}
		}
		err = directoriesFile.Close()
		if err != nil {
			log.Errorf("Could not close directories.wxi: %v", packageName, err)
			os.Exit(1)
		}
		for _, line := range directoryRefsFileContent {
			if _, err := directoryRefsFile.WriteString(line + "\n"); err != nil {
				log.Errorf("Could not write directory_refs.wxi: %v", packageName, err)
				os.Exit(1)
			}
		}
		err = directoryRefsFile.Close()
		if err != nil {
			log.Errorf("Could not close directory_refs.wxi: %v", packageName, err)
			os.Exit(1)
		}
		for _, line := range componentRefsFileContent {
			if _, err := componentRefsFile.WriteString(line + "\n"); err != nil {
				log.Errorf("Could not write component_refs.wxi: %v", packageName, err)
				os.Exit(1)
			}
		}
		err = componentRefsFile.Close()
		if err != nil {
			log.Errorf("Could not close component_refs.wxi: %v", packageName, err)
			os.Exit(1)
		}
	},
}

func windowsMsiProcessFiles(path string) {
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
			windowsMsiProcessFiles(p)
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
