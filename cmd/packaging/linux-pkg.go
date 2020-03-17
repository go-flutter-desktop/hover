package packaging

// LinuxPkgTask packaging for linux as pacman pkg
var LinuxPkgTask = &packagingTask{
	packagingFormatName: "linux-pkg",
	templateFiles: map[string]string{
		"linux-pkg/PKGBUILD.tmpl.tmpl": "PKGBUILD.tmpl",
		"linux/bin.tmpl":               "src/usr/bin/{{.projectName}}",
		"linux/app.desktop.tmpl":       "src/usr/share/applications/{{.projectName}}.desktop",
	},
	executableFiles: []string{
		"src/usr/bin/{{.projectName}}",
		"src/usr/share/applications/{{.projectName}}.desktop",
	},
	linuxDesktopFileExecutablePath: "/usr/lib/{{.projectName}}/{{.projectName}}",
	linuxDesktopFileIconPath:       "/usr/lib/{{.projectName}}/assets/icon.png",
	buildOutputDirectory:           "src/usr/lib/{{.projectName}}",
	packagingScriptTemplate:        "makepkg && mv {{.projectName}}-{{.version}}-{{.release}}-x86_64.pkg.tar.xz {{.projectName}}-{{.version}}.pkg.tar.xz",
	outputFileExtension:            "pkg.tar.xz",
}
