package packaging

// DarwinBundlePackagingTask packaging for darwin as bundle
var DarwinBundlePackagingTask = &packagingTask{
	packagingFormatName: "darwin-bundle",
	templateFiles: map[string]string{
		"darwin-bundle/Info.plist.tmpl.tmpl": "{{.projectName}}.app/Contents/Info.plist.tmpl",
	},
	executableFiles: []string{},
	dockerfileContent: []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install icnsutils -y",
	},
	buildOutputDirectory:      "{{.projectName}}.app/Contents/MacOS",
	packagingScriptTemplate:   "mkdir -p {{.projectName}}.app/Contents/Resources && png2icns {{.projectName}}.app/Contents/Resources/icon.icns {{.projectName}}.app/Contents/MacOS/assets/icon.png",
	outputFileExtension:       "app",
	outputFileContainsVersion: false,
}
