package packaging

// LinuxAppImagePackagingTask packaging for linux as AppImage
var LinuxAppImagePackagingTask = &packagingTask{
	packagingFormatName: "linux-appimage",
	templateFiles: map[string]string{
		"linux-appimage/AppRun.tmpl": "AppRun",
		"linux/app.desktop.tmpl":     "{{.projectName}}.desktop",
	},
	executableFiles: []string{
		"AppRun",
		"{{.projectName}}.desktop",
	},
	linuxDesktopFileIconPath: "/build/assets/icon",
	dockerfileContent: []string{
		"FROM ubuntu:bionic",
		"WORKDIR /opt",
		"RUN apt-get update && \\",
		"apt-get install libglib2.0-0 curl file -y",
		"RUN curl -LO https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage && \\",
		"chmod a+x appimagetool-x86_64.AppImage && \\",
		"./appimagetool-x86_64.AppImage --appimage-extract && \\",
		"mv squashfs-root appimagetool && \\",
		"rm appimagetool-x86_64.AppImage",
		"ENV PATH=/opt/appimagetool/usr/bin:$PATH",
	},
	buildOutputDirectory:    "build",
	packagingScriptTemplate: "appimagetool . && mv {{.projectName}}-x86_64.AppImage {{.projectName}}-{{.version}}.AppImage",
	outputFileExtension:     "AppImage",
}
