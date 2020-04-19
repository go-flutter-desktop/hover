package packaging

// LinuxAppImageTask packaging for linux as AppImage
var LinuxAppImageTask = &packagingTask{
	packagingFormatName: "linux-appimage",
	templateFiles: map[string]string{
		"linux-appimage/AppRun.tmpl": "AppRun",
		"linux/app.desktop.tmpl":     "{{.projectName}}.desktop",
	},
	executableFiles: []string{
		"AppRun",
		"{{.projectName}}.desktop",
	},
	linuxDesktopFileIconPath:  "/build/assets/icon",
	buildOutputDirectory:      "build",
	packagingScriptTemplate:   "appimagetool . && mv {{.projectName}}-x86_64.AppImage {{.projectName}}-{{.version}}.AppImage",
	outputFileExtension:       "AppImage",
	outputFileContainsVersion: true,
}
