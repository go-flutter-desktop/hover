package packaging

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	ico "github.com/Kodeworks/golang-image-ico"

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
	flutterBuildOutputDirectory: "build",
	packagingFunction: func(tmpPath, applicationName, strippedApplicationName, packageName, executableName, version, release string) (string, error) {
		outputFileName := fmt.Sprintf("%s %s.msi", applicationName, version)
		iconPngFile, err := os.Open(filepath.Join(tmpPath, "build", "assets", "icon.png"))
		if err != nil {
			return "", err
		}
		pngImage, err := png.Decode(iconPngFile)
		if err != nil {
			return "", err
		}
		// We can't defer it, because windows reports that the file is used by another program
		err = iconPngFile.Close()
		if err != nil {
			return "", err
		}
		iconIcoFile, err := os.Create(filepath.Join(tmpPath, "build", "assets", "icon.ico"))
		if err != nil {
			return "", err
		}
		err = ico.Encode(iconIcoFile, pngImage)
		if err != nil {
			return "", err
		}
		// We can't defer it, because windows reports that the file is used by another program
		err = iconIcoFile.Close()
		if err != nil {
			return "", err
		}
		switch runtime.GOOS {
		case "windows":
			cmdCandle := exec.Command("candle", fmt.Sprintf("%s.wxs", packageName))
			cmdCandle.Dir = tmpPath
			cmdCandle.Stdout = os.Stdout
			cmdCandle.Stderr = os.Stderr
			err = cmdCandle.Run()
			if err != nil {
				return "", err
			}
			cmdLight := exec.Command("light", fmt.Sprintf("%s.wixobj", packageName), "-sval")
			cmdLight.Dir = tmpPath
			cmdLight.Stdout = os.Stdout
			cmdLight.Stderr = os.Stderr
			err = cmdLight.Run()
			if err != nil {
				return "", err
			}
			err = os.Rename(filepath.Join(tmpPath, fmt.Sprintf("%s.msi", packageName)), filepath.Join(tmpPath, outputFileName))
			if err != nil {
				return "", err
			}
		case "linux":
			cmdWixl := exec.Command("wixl", "-v", fmt.Sprintf("%s.wxs", packageName), "-o", outputFileName)
			cmdWixl.Dir = tmpPath
			cmdWixl.Stdout = os.Stdout
			cmdWixl.Stderr = os.Stderr
			err = cmdWixl.Run()
			if err != nil {
				return "", err
			}
		default:
			panic("should be unreachable")
		}
		return outputFileName, nil
	},
	requiredTools: map[string][]string{
		"windows": {"candle", "light"},
		"linux":   {"wixl"},
	},
	generateInitFiles: func(packageName, path string) {
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			log.Errorf("Failed to generate GUID: %v", err)
			os.Exit(1)
		}
		upgradeCode := strings.ToUpper(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]))
		err = ioutil.WriteFile(filepath.Join(path, "upgrade-code.txt"), []byte(fmt.Sprintf("%s\n# This GUID is your upgrade code and ensures that you can properly update your app.\n# Don't change it.", upgradeCode)), 0755)
		if err != nil {
			log.Errorf("Failed to create `upgrade-code.txt` file: %v", err)
			os.Exit(1)
		}
	},
	extraTemplateData: func(packageName, path string) map[string]string {
		data, err := ioutil.ReadFile(filepath.Join(path, "upgrade-code.txt"))
		if err != nil {
			log.Errorf("Failed to read `upgrade-code.txt`: %v", err)
			if os.IsNotExist(err) {
				log.Errorf("Please re-init windows-msi to create an `upgrade-code.txt`")
				log.Errorf("or put a GUID from https://www.guidgen.com/ into a new `upgrade-code.txt` file.")
			}
			os.Exit(1)
		}
		guid := strings.Split(string(data), "\n")[0]
		return map[string]string{
			"upgradeCode":   guid,
			"pathSeparator": string(os.PathSeparator),
		}
	},
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
		directoriesFileContent = append(directoriesFileContent, "<Include>")
		directoryRefsFileContent = append(directoryRefsFileContent, "<Include>")
		componentRefsFileContent = append(componentRefsFileContent, "<Include>")
		windowsMsiProcessFiles(filepath.Join(tmpPath, "build", "flutter_assets"))
		directoriesFileContent = append(directoriesFileContent, "</Include>")
		directoryRefsFileContent = append(directoryRefsFileContent, "</Include>")
		componentRefsFileContent = append(componentRefsFileContent, "</Include>")

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
	pathSeparator := string(os.PathSeparator)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Errorf("Failed to read directory %s: %v", path, err)
		os.Exit(1)
	}

	for _, f := range files {
		p := filepath.Join(path, f.Name())
		relativePath := strings.Split(strings.Split(p, "flutter_assets"+pathSeparator)[1], pathSeparator)
		id := hashSha1(strings.Join(relativePath, ""))
		if f.IsDir() {
			directoriesFileContent = append(directoriesFileContent,
				fmt.Sprintf(`<Directory Id="FLUTTERASSETSDIRECTORY_%s" Name="%s">`, id, f.Name()),
			)
			windowsMsiProcessFiles(p)
			directoriesFileContent = append(directoriesFileContent,
				"</Directory>",
			)
		} else {
			if len(relativePath) > 1 {
				directoryRefsFileContent = append(directoryRefsFileContent,
					fmt.Sprintf(`<DirectoryRef Id="FLUTTERASSETSDIRECTORY_%s">`, hashSha1(strings.Join(relativePath[:len(relativePath)-1], ""))),
				)
			} else {
				directoryRefsFileContent = append(directoryRefsFileContent,
					`<DirectoryRef Id="FLUTTERASSETSDIRECTORY">`,
				)
			}
			fileSource := filepath.Join("build", "flutter_assets", strings.Join(relativePath, pathSeparator))
			directoryRefsFileContent = append(directoryRefsFileContent,
				fmt.Sprintf(`<Component Id="flutter_assets_%s" Guid="*">`, id),
				fmt.Sprintf(`<File Id="flutter_assets_%s" Source="%s" KeyPath="yes"/>`, id, fileSource),
				"</Component>",
				"</DirectoryRef>",
			)
			componentRefsFileContent = append(componentRefsFileContent,
				fmt.Sprintf(`<ComponentRef Id="flutter_assets_%s"/>`, id),
			)
		}
	}
}

func hashSha1(content string) string {
	h := sha1.New()
	h.Write([]byte(content))
	sha := h.Sum(nil)
	return hex.EncodeToString(sha)
}
