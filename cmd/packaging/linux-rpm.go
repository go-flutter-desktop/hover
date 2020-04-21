package packaging

// LinuxRpmTask packaging for linux as rpm
var LinuxRpmTask = &packagingTask{
	packagingFormatName: "linux-rpm",
	templateFiles: map[string]string{
		"linux-rpm/app.spec.tmpl": "SPECS/{{.packageName}}.spec.tmpl",
		"linux/bin.tmpl":          "BUILDROOT/{{.packageName}}-{{.version}}-{{.release}}.x86_64/usr/bin/{{.executableName}}.tmpl",
		"linux/app.desktop.tmpl":  "BUILDROOT/{{.packageName}}-{{.version}}-{{.release}}.x86_64/usr/share/applications/{{.executableName}}.desktop.tmpl",
	},
	executableFiles: []string{
		"BUILDROOT/{{.packageName}}-{{.version}}-{{.release}}.x86_64/usr/bin/{{.executableName}}",
		"BUILDROOT/{{.packageName}}-{{.version}}-{{.release}}.x86_64/usr/share/applications/{{.executableName}}.desktop",
	},
	linuxDesktopFileExecutablePath: "/usr/lib/{{.packageName}}/{{.executableName}}",
	linuxDesktopFileIconPath:       "/usr/lib/{{.packageName}}/assets/icon.png",
	buildOutputDirectory:           "BUILD/{{.packageName}}-{{.version}}-{{.release}}.x86_64/usr/lib/{{.packageName}}",
	packagingScriptTemplate:        "rpmbuild --define \"_topdir $(pwd)\" --define \"_unpackaged_files_terminate_build 0\" -ba ./SPECS/{{.packageName}}.spec && mv RPMS/x86_64/{{.packageName}}-{{.version}}-{{.release}}.x86_64.rpm {{.packageName}}-{{.version}}.rpm",
	outputFileExtension:            "rpm",
	outputFileContainsVersion:      true,
	outputFileUsesApplicationName:  false,
}
