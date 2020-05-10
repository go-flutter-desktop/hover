package packaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DarwinBundleTask packaging for darwin as bundle
var DarwinBundleTask = &packagingTask{
	packagingFormatName: "darwin-bundle",
	templateFiles: map[string]string{
		"darwin-bundle/Info.plist.tmpl": "{{.applicationName}} {{.version}}.app/Contents/Info.plist.tmpl",
	},
	executableFiles:             []string{},
	flutterBuildOutputDirectory: "{{.applicationName}} {{.version}}.app/Contents/MacOS",
	packagingFunction: func(tmpPath, applicationName, strippedApplicationName, packageName, executableName, version, release string) (string, error) {
		outputFileName := fmt.Sprintf("%s %s.app", applicationName, version)
		err := os.MkdirAll(filepath.Join(tmpPath, outputFileName, "Contents", "Resources"), 0755)
		if err != nil {
			return "", err
		}
		cmdPng2icns := exec.Command("png2icns", filepath.Join(outputFileName, "Contents", "Resources", "icon.icns"), filepath.Join(outputFileName, "Contents", "MacOS", "assets", "icon.png"))
		cmdPng2icns.Dir = tmpPath
		cmdPng2icns.Stdout = os.Stdout
		cmdPng2icns.Stderr = os.Stderr
		err = cmdPng2icns.Run()
		if err != nil {
			return "", err
		}
		return outputFileName, nil
	},
	requiredTools: map[string][]string{
		"linux": {"png2icns"},
	},
}
