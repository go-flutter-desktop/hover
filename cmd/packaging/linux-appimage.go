package packaging

import (
	"fmt"
	"os"
	"os/exec"
)

// LinuxAppImageTask packaging for linux as AppImage
var LinuxAppImageTask = &packagingTask{
	packagingFormatName: "linux-appimage",
	templateFiles: map[string]string{
		"linux-appimage/AppRun.tmpl": "AppRun.tmpl",
		"linux/app.desktop.tmpl":     "{{.executableName}}.desktop.tmpl",
	},
	executableFiles: []string{
		"AppRun",
		"{{.executableName}}.desktop",
	},
	linuxDesktopFileIconPath:    "/build/assets/icon",
	flutterBuildOutputDirectory: "build",
	packagingFunction: func(tmpPath, applicationName, packageName, executableName, version, release string) (string, error) {
		cmdAppImageTool := exec.Command("appimagetool", ".")
		cmdAppImageTool.Dir = tmpPath
		cmdAppImageTool.Stdout = os.Stdout
		cmdAppImageTool.Stderr = os.Stderr
		err := cmdAppImageTool.Run()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s-x86_64.AppImage", executableName), nil
	},
	requiredTools: map[string][]string{
		"linux": {"appimagetool"},
	},
}
