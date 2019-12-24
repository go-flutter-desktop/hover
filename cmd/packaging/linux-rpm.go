package packaging

// LinuxRpmTask packaging for linux as rpm
var LinuxRpmTask = &packagingTask{
	packagingFormatName: "linux-rpm",
	templateFiles: map[string]string{
		"linux-rpm/app.spec.tmpl.tmpl": "rpmbuild/SPECS/{{.projectName}}.spec.tmpl",
		"linux/bin.tmpl":               "rpmbuild/BUILDROOT/{{.projectName}}{{`-{{.version}}-{{.version}}`}}.x86_64/usr/bin/{{.projectName}}",
		"linux/app.desktop.tmpl":       "rpmbuild/BUILDROOT/{{.projectName}}{{`-{{.version}}-{{.version}}`}}.x86_64/usr/share/applications/{{.projectName}}.desktop",
	},
	executableFiles: []string{
		"rpmbuild/BUILDROOT/{{.projectName}}-{{.version}}-{{.version}}.x86_64/usr/bin/{{.projectName}}",
		"rpmbuild/BUILDROOT/{{.projectName}}-{{.version}}-{{.version}}.x86_64/usr/share/applications/{{.projectName}}.desktop",
	},
	linuxDesktopFileExecutablePath: "/usr/lib/{{.projectName}}/{{.projectName}}",
	linuxDesktopFileIconPath:       "/usr/lib/{{.projectName}}/assets/icon.png",
	buildOutputDirectory:           "rpmbuild/BUILDROOT/{{.projectName}}-{{.version}}-{{.version}}.x86_64/usr/lib/{{.projectName}}",
	packagingScriptTemplate:        "rpmbuild --define '_topdir /app/rpmbuild' -ba /app/rpmbuild/SPECS/{{.projectName}}.spec && rm /root/.rpmdb -r && mv rpmbuild/RPMS/x86_64/{{.projectName}}-{{.version}}-{{.version}}.x86_64.rpm {{.projectName}}-{{.version}}.rpm",
	outputFileExtension:            "rpm",
}
