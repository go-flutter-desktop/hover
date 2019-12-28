package packaging

// LinuxDebPackagingTask packaging for linux as deb
var LinuxDebPackagingTask = &packagingTask{
	packagingFormatName: "linux-deb",
	templateFiles: map[string]string{
		"linux-deb/control.tmpl.tmpl": "DEBIAN/control.tmpl",
		"linux/bin.tmpl":              "usr/bin/{{.projectName}}",
		"linux/app.desktop.tmpl":      "usr/share/applications/{{.projectName}}.desktop",
	},
	executableFiles: []string{
		"usr/bin/{{.projectName}}",
		"usr/share/applications/{{.projectName}}.desktop",
	},
	linuxDesktopFileExecutablePath: "/usr/lib/{{.projectName}}/{{.projectName}}",
	linuxDesktopFileIconPath:       "/usr/lib/{{.projectName}}/assets/icon.png",
	dockerfileContent: []string{
		"FROM ubuntu:bionic",
	},
	buildOutputDirectory:    "usr/lib/{{.projectName}}",
	packagingScriptTemplate: "dpkg-deb --build . {{.projectName}}-{{.version}}.deb",
	outputFileExtension:     "deb",
}
