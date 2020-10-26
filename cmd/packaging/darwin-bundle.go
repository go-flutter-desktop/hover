package packaging

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/JackMordaunt/icns"
)

// DarwinBundleTask packaging for darwin as bundle
var DarwinBundleTask = &packagingTask{
	packagingFormatName: "darwin-bundle",
	templateFiles: map[string]string{
		"darwin-bundle/Info.plist.tmpl": "{{.applicationName}} {{.version}}.app/Contents/Info.plist.tmpl",
	},
	executableFiles:             []string{},
	flutterBuildOutputDirectory: "{{.applicationName}} {{.version}}.app/Contents/MacOS",
	packagingFunction: func(tmpPath, applicationName, packageName, executableName, version, release string) (string, error) {
		outputFileName := fmt.Sprintf("%s %s.app", applicationName, version)
		err := os.MkdirAll(filepath.Join(tmpPath, outputFileName, "Contents", "Resources"), 0755)
		if err != nil {
			return "", err
		}

		pngFile, err := os.Open(filepath.Join(tmpPath, outputFileName, "Contents", "MacOS", "assets", "icon.png"))
		if err != nil {
			return "", err
		}
		defer pngFile.Close()

		srcImg, _, err := image.Decode(pngFile)
		if err != nil {
			return "", err
		}

		icnsFile, err := os.Create(filepath.Join(tmpPath, outputFileName, "Contents", "Resources", "icon.icns"))
		if err != nil {
			return "", err
		}
		defer icnsFile.Close()

		err = icns.Encode(icnsFile, srcImg)
		if err != nil {
			return "", err
		}

		return outputFileName, nil
	},
	requiredTools: map[string][]string{
		"linux":   {},
		"darwin":  {},
		"windows": {},
	},
}
