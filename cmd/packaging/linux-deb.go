package packaging

import (
	"fmt"
	"os"
	"os/exec"
)

// LinuxDebTask packaging for linux as deb
var LinuxDebTask = &packagingTask{
	packagingFormatName: "linux-deb",
	templateFiles: map[string]string{
		"linux-deb/control.tmpl": "DEBIAN/control.tmpl",
		"linux/bin.tmpl":         "usr/bin/{{.executableName}}.tmpl",
		"linux/app.desktop.tmpl": "usr/share/applications/{{.executableName}}.desktop.tmpl",
	},
	executableFiles: []string{
		"usr/bin/{{.executableName}}",
		"usr/share/applications/{{.executableName}}.desktop",
	},
	linuxDesktopFileExecutablePath: "/usr/lib/{{.packageName}}/{{.executableName}}",
	linuxDesktopFileIconPath:       "/usr/lib/{{.packageName}}/assets/icon.png",
	flutterBuildOutputDirectory:    "usr/lib/{{.packageName}}",
	packagingFunction: func(tmpPath, applicationName, packageName, executableName, version, release string) (string, error) {
		outputFileName := fmt.Sprintf("%s_%s_amd64.deb", packageName, version)
		cmdDpkgDeb := exec.Command("dpkg-deb", "--build", ".", outputFileName)
		cmdDpkgDeb.Dir = tmpPath
		cmdDpkgDeb.Stdout = os.Stdout
		cmdDpkgDeb.Stderr = os.Stderr
		err := cmdDpkgDeb.Run()
		if err != nil {
			return "", err
		}
		return outputFileName, nil
	},
	requiredTools: map[string]map[string]string{
		"linux": {
			"makepkg": "You need to be on Debian, Ubuntu or another distro that uses apt/dpkg as package manager to use this. Installing dpkg on other distros is hard and dangerous.",
		},
	},
}
