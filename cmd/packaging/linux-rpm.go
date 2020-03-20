package packaging

// LinuxRpmPackagingTask packaging for linux as rpm
var LinuxRpmPackagingTask = &packagingTask{
	packagingFormatName: "linux-rpm",
	templateFiles: map[string]string{
		"linux-rpm/app.spec.tmpl.tmpl": "rpmbuild/SPECS/{{.projectName}}.spec.tmpl",
		"linux/bin.tmpl":               "rpmbuild/BUILDROOT/{{.projectName}}{{`-{{.version}}-{{.release}}`}}.x86_64/usr/bin/{{.projectName}}",
		"linux/app.desktop.tmpl":       "rpmbuild/BUILDROOT/{{.projectName}}{{`-{{.version}}-{{.release}}`}}.x86_64/usr/share/applications/{{.projectName}}.desktop",
	},
	executableFiles: []string{
		"rpmbuild/BUILDROOT/{{.projectName}}-{{.version}}-{{.release}}.x86_64/usr/bin/{{.projectName}}",
		"rpmbuild/BUILDROOT/{{.projectName}}-{{.version}}-{{.release}}.x86_64/usr/share/applications/{{.projectName}}.desktop",
	},
	linuxDesktopFileExecutablePath: "/usr/lib/{{.projectName}}/{{.projectName}}",
	linuxDesktopFileIconPath:       "/usr/lib/{{.projectName}}/assets/icon.png",
	dockerfileContent: []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install rpm file -y",
	},
	buildOutputDirectory:      "rpmbuild/BUILDROOT/{{.projectName}}-{{.version}}-{{.release}}.x86_64/usr/lib/{{.projectName}}",
	packagingScriptTemplate:   "rpmbuild --define '_topdir /app/rpmbuild' -ba /app/rpmbuild/SPECS/{{.projectName}}.spec && rm /root/.rpmdb -r && mv rpmbuild/RPMS/x86_64/{{.projectName}}-{{.version}}-{{.release}}.x86_64.rpm {{.projectName}}-{{.version}}.rpm",
	outputFileExtension:       "rpm",
	outputFileContainsVersion: true,
}
