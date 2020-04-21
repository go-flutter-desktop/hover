package packaging

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
	buildOutputDirectory:           "usr/lib/{{.packageName}}",
	packagingScriptTemplate:        "dpkg-deb --build . {{.packageName}}-{{.version}}.deb",
	outputFileExtension:            "deb",
	outputFileContainsVersion:      true,
	outputFileUsesApplicationName:  false,
}
