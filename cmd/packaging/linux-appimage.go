package packaging

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
	linuxDesktopFileIconPath:      "/build/assets/icon",
	buildOutputDirectory:          "build",
	packagingScriptTemplate:       "appimagetool . && mv -n {{.executableName}}-x86_64.AppImage {{.packageName}}-{{.version}}.AppImage",
	outputFileExtension:           "AppImage",
	outputFileContainsVersion:     true,
	outputFileUsesApplicationName: false,
}
