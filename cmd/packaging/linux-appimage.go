package packaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	copy "github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/log"
)

// LinuxAppImageTask packaging for linux as AppImage
var LinuxAppImageTask = &packagingTask{
	packagingFormatName: "linux-appimage",
	templateFiles: map[string]string{
		"linux-appimage/AppRun.tmpl": "AppRun.tmpl",
		"linux/app.desktop.tmpl":     "{{.strippedApplicationName}}.desktop.tmpl",
	},
	executableFiles: []string{
		".",
		"AppRun",
		"{{.strippedApplicationName}}.desktop",
	},
	linuxDesktopFileIconPath:    "{{.strippedApplicationName}}",
	flutterBuildOutputDirectory: "build",
	packagingFunction: func(tmpPath, applicationName, strippedApplicationName, packageName, executableName, version, release string) (string, error) {
		sourceIconPath := filepath.Join(tmpPath, "build", "assets", "icon.png")
		iconDir := filepath.Join(tmpPath, "usr", "share", "icons", "hicolor", "256x256", "apps")
		if _, err := os.Stat(iconDir); os.IsNotExist(err) {
			err = os.MkdirAll(iconDir, 0755)
			if err != nil {
				log.Errorf("Failed to create icon dir: %v", err)
				os.Exit(1)
			}
		}
		err := copy.Copy(sourceIconPath, filepath.Join(tmpPath, fmt.Sprintf("%s.png", strippedApplicationName)))
		if err != nil {
			log.Errorf("Failed to copy icon root dir: %v", err)
			os.Exit(1)
		}
		err = copy.Copy(sourceIconPath, filepath.Join(iconDir, fmt.Sprintf("%s.png", strippedApplicationName)))
		if err != nil {
			log.Errorf("Failed to copy icon dir: %v", err)
			os.Exit(1)
		}
		cmdAppImageTool := exec.Command("appimagetool", ".")
		cmdAppImageTool.Dir = tmpPath
		cmdAppImageTool.Stdout = os.Stdout
		cmdAppImageTool.Stderr = os.Stderr
		cmdAppImageTool.Env = append(
			os.Environ(),
			"ARCH=x86_64",
			fmt.Sprintf("VERSION=%s", version),
		)
		err = cmdAppImageTool.Run()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s-%s-x86_64.AppImage", strippedApplicationName, version), nil
	},
	requiredTools: map[string][]string{
		"linux": {"appimagetool"},
	},
}
