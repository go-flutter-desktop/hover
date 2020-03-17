package packaging

// DarwinDmgTask packaging for darwin as dmg
var DarwinDmgTask = &packagingTask{
	packagingFormatName: "darwin-dmg",
	dependsOn: map[*packagingTask]string{
		DarwinBundleTask: "dmgdir",
	},
	packagingScriptTemplate: "ln -sf /Applications dmgdir/Applications && genisoimage -V '{{.projectName}}' -D -R -apple -no-pad -o '{{.projectName}}-{{.version}}.dmg' dmgdir",
	outputFileExtension:     "dmg",
}
