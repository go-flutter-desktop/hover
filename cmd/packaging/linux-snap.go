package packaging

// LinuxSnapPackagingTask packaging for linux as snao
var LinuxSnapPackagingTask = &packagingTask{
	packagingFormatName: "linux-snap",
	templateFiles: map[string]string{
		"linux-snap/snapcraft.yaml.tmpl.tmpl": "snap/snapcraft.yaml.tmpl",
		"linux/app.desktop.tmpl":              "snap/local/{{.projectName}}.desktop",
	},
	linuxDesktopFileExecutablePath: "/{{.projectName}}",
	linuxDesktopFileIconPath:       "/icon.png",
	dockerfileContent: []string{
		"FROM snapcore/snapcraft",
	},
	buildOutputDirectory:    "build",
	packagingScriptTemplate: "snapcraft && mv {{.strippedProjectName}}_{{.version}}_{{.arch}}.snap {{.projectName}}-{{.version}}.snap",
	outputFileExtension:     "snap",
}
