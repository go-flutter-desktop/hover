package packaging

// DarwinBundleTask packaging for darwin as bundle
var DarwinBundleTask = &packagingTask{
	packagingFormatName: "darwin-bundle",
	templateFiles: map[string]string{
		"darwin-bundle/Info.plist.tmpl": "{{.applicationName}} {{.version}}.app/Contents/Info.plist.tmpl",
	},
	executableFiles:               []string{},
	buildOutputDirectory:          "{{.applicationName}} {{.version}}.app/Contents/MacOS",
	packagingScriptTemplate:       "mkdir -p \"{{.applicationName}} {{.version}}.app/Contents/Resources\" && png2icns \"{{.applicationName}} {{.version}}.app/Contents/Resources/icon.icns\" \"{{.applicationName}} {{.version}}.app/Contents/MacOS/assets/icon.png\"",
	outputFileExtension:           "app",
	outputFileContainsVersion:     true,
	outputFileUsesApplicationName: true,
}
