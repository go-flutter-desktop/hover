package packaging

// DarwinDmgTask packaging for darwin as dmg
var DarwinDmgTask = &packagingTask{
	packagingFormatName: "darwin-dmg",
	dependsOn: map[*packagingTask]string{
		DarwinBundleTask: "dmgdir",
	},
	packagingScriptTemplate:       "ln -sf /Applications dmgdir/Applications && genisoimage -V {{.packageName}} -D -R -apple -no-pad -o \"{{.applicationName}} {{.version}}.dmg\" dmgdir",
	outputFileExtension:           "dmg",
	outputFileContainsVersion:     true,
	outputFileUsesApplicationName: true,
	skipAssertInitialized:         true,
}
