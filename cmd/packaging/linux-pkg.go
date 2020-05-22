package packaging

import (
	"fmt"
	"os"
	"os/exec"
)

// LinuxPkgTask packaging for linux as pacman pkg
var LinuxPkgTask = &packagingTask{
	packagingFormatName: "linux-pkg",
	templateFiles: map[string]string{
		"linux-pkg/PKGBUILD.tmpl": "PKGBUILD.tmpl",
		"linux/bin.tmpl":          "src/usr/bin/{{.executableName}}.tmpl",
		"linux/app.desktop.tmpl":  "src/usr/share/applications/{{.executableName}}.desktop.tmpl",
	},
	executableFiles: []string{
		"src/usr/bin/{{.executableName}}",
		"src/usr/share/applications/{{.executableName}}.desktop",
	},
	linuxDesktopFileExecutablePath: "/usr/lib/{{.packageName}}/{{.executableName}}",
	linuxDesktopFileIconPath:       "/usr/lib/{{.packageName}}/assets/icon",
	flutterBuildOutputDirectory:    "src/usr/lib/{{.packageName}}",
	packagingFunction: func(tmpPath, applicationName, packageName, executableName, version, release string) (string, error) {
		cmdMakepkg := exec.Command("makepkg")
		cmdMakepkg.Dir = tmpPath
		cmdMakepkg.Stdout = os.Stdout
		cmdMakepkg.Stderr = os.Stderr
		err := cmdMakepkg.Run()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s-%s-%s-x86_64.pkg.tar.xz", packageName, version, release), nil
	},
	requiredTools: map[string][]string{
		"linux": {"makepkg"},
	},
}
