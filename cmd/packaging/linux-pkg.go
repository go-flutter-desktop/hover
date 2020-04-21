package packaging

// LinuxPkgTask packaging for linux as pacman pkg
var LinuxPkgTask = &packagingTask{
	packagingFormatName: "linux-pkg",
	templateFiles: map[string]string{
		"linux-pkg/PKGBUILD.tmpl": "PKGBUILD.tmpl",
		"linux/bin.tmpl":          "src/usr/bin/{{.executableName}}.tmpl",
		"linux/app.desktop.tmpl":  "src/usr/share/applications/{{.executableName}}.desktop.tmpl",
	},
	executableFiles: []string{
		"src/usr/bin/{{.executableName}}",
		"src/usr/share/applications/{{.executableName}}.desktop",
	},
	linuxDesktopFileExecutablePath: "/usr/lib/{{.packageName}}/{{.executableName}}",
	linuxDesktopFileIconPath:       "/usr/lib/{{.packageName}}/assets/icon.png",
	buildOutputDirectory:           "src/usr/lib/{{.packageName}}",
	packagingScriptTemplate:        "makepkg && mv -p {{.packageName}}-{{.version}}-{{.release}}-x86_64.pkg.tar.xz {{.packageName}}-{{.version}}.pkg.tar.xz",
	outputFileExtension:            "pkg.tar.xz",
	outputFileContainsVersion:      true,
	outputFileUsesApplicationName:  false,
}
