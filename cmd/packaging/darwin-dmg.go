package packaging

// DarwinDmgPackagingTask packaging for darwin as dmg
var DarwinDmgPackagingTask = &packagingTask{
	packagingFormatName: "darwin-dmg",
	dependsOn: map[*packagingTask]string{
		DarwinBundlePackagingTask: "dmgdir",
	},
	dockerfileContent: []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install genisoimage -y ",
	},
	packagingScriptTemplate:   "ln -sf /Applications dmgdir/Applications && genisoimage -V '{{.projectName}}' -D -R -apple -no-pad -o '{{.projectName}}-{{.version}}.dmg' dmgdir",
	outputFileExtension:       "dmg",
	outputFileContainsVersion: true,
}
