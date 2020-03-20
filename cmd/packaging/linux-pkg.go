package packaging

// LinuxPkgPackagingTask packaging for linux as pacman pkg
var LinuxPkgPackagingTask = &packagingTask{
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
	dockerfileContent: []string{
		"FROM archlinux",
		"RUN pacman -Sy fakeroot base-devel --noconfirm",
		"RUN useradd --no-create-home --shell=/bin/false build && usermod -L build",
		"USER build",
	},
	buildOutputDirectory:      "src/usr/lib/{{.projectName}}",
	packagingScriptTemplate:   "makepkg && mv {{.projectName}}-{{.version}}-{{.release}}-x86_64.pkg.tar.xz {{.projectName}}-{{.version}}.pkg.tar.xz",
	outputFileExtension:       "pkg.tar.xz",
	outputFileContainsVersion: true,
}
