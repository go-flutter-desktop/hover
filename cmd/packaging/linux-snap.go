package packaging

// LinuxSnapTask packaging for linux as snap
var LinuxSnapTask = &packagingTask{
	packagingFormatName: "linux-snap",
	templateFiles: map[string]string{
		"linux-snap/snapcraft.yaml.tmpl": "snap/snapcraft.yaml.tmpl",
		"linux/app.desktop.tmpl":         "snap/local/{{.executableName}}.desktop.tmpl",
	},
	linuxDesktopFileExecutablePath: "/{{.executableName}}",
	linuxDesktopFileIconPath:       "/icon.png",
	buildOutputDirectory:           "build",
	packagingScriptTemplate:        "snapcraft && mv -n {{.packageName}}_{{.version}}_{{.arch}}.snap {{.packageName}}-{{.version}}.snap",
	outputFileExtension:            "snap",
	outputFileContainsVersion:      true,
	outputFileUsesApplicationName:  false,
}
