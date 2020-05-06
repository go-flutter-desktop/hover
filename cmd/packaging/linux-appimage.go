package packaging

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
	linuxDesktopFileIconPath:      "{{.strippedApplicationName}}",
	buildOutputDirectory:          "build",
	packagingScriptTemplate:       "cp build/assets/icon.png {{.strippedApplicationName}}.png && mkdir -p usr/share/icons/hicolor/256x256/apps && cp build/assets/icon.png usr/share/icons/hicolor/256x256/apps/{{.strippedApplicationName}}.png && ARCH=x86_64 VERSION={{.version}} appimagetool . && mv -n {{.strippedApplicationName}}-{{.version}}-x86_64.AppImage {{.packageName}}-{{.version}}.AppImage",
	outputFileExtension:           "AppImage",
	outputFileContainsVersion:     true,
	outputFileUsesApplicationName: false,
}
