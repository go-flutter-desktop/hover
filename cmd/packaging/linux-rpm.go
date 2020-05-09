package packaging

import (
	"fmt"
	"os"
	"os/exec"
)

// LinuxRpmTask packaging for linux as rpm
var LinuxRpmTask = &packagingTask{
	packagingFormatName: "linux-rpm",
	templateFiles: map[string]string{
		"linux-rpm/app.spec.tmpl": "SPECS/{{.packageName}}.spec.tmpl",
		"linux/bin.tmpl":          "BUILDROOT/{{.packageName}}-{{.version}}-{{.release}}.x86_64/usr/bin/{{.executableName}}.tmpl",
		"linux/app.desktop.tmpl":  "BUILDROOT/{{.packageName}}-{{.version}}-{{.release}}.x86_64/usr/share/applications/{{.executableName}}.desktop.tmpl",
	},
	executableFiles: []string{
		"BUILDROOT/{{.packageName}}-{{.version}}-{{.release}}.x86_64/usr/bin/{{.executableName}}",
		"BUILDROOT/{{.packageName}}-{{.version}}-{{.release}}.x86_64/usr/share/applications/{{.executableName}}.desktop",
	},
	linuxDesktopFileExecutablePath: "/usr/lib/{{.packageName}}/{{.executableName}}",
	linuxDesktopFileIconPath:       "/usr/lib/{{.packageName}}/assets/icon.png",
	flutterBuildOutputDirectory:    "BUILD/{{.packageName}}-{{.version}}-{{.release}}.x86_64/usr/lib/{{.packageName}}",
	packagingFunction: func(tmpPath, applicationName, strippedApplicationName, packageName, executableName, version, release string) (string, error) {
		cmdRpmbuild := exec.Command("rpmbuild", "--define", fmt.Sprintf("_topdir %s", tmpPath), "--define", "_unpackaged_files_terminate_build 0", "-ba", fmt.Sprintf("./SPECS/%s.spec", packageName))
		cmdRpmbuild.Dir = tmpPath
		cmdRpmbuild.Stdout = os.Stdout
		cmdRpmbuild.Stderr = os.Stderr
		err := cmdRpmbuild.Run()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("RPMS/x86_64/%s-%s-%s.x86_64.rpm", packageName, version, release), nil
	},
	requiredTools: map[string][]string{
		"linux": {"rpmbuild"},
	},
}
