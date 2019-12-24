package packaging

// DarwinPkgTask packaging for darwin as pkg
var DarwinPkgTask = &packagingTask{
	packagingFormatName: "darwin-pkg",
	dependsOn: map[*packagingTask]string{
		DarwinBundleTask: "flat/root/Applications",
	},
	templateFiles: map[string]string{
		"darwin-pkg/PackageInfo.tmpl.tmpl":  "flat/base.pkg/PackageInfo.tmpl",
		"darwin-pkg/Distribution.tmpl.tmpl": "flat/Distribution.tmpl",
	},
	packagingScriptTemplate: "(cd flat/root && find . | cpio -o --format odc --owner 0:80 | gzip -c ) > flat/base.pkg/Payload && mkbom -u 0 -g 80 flat/root flat/base.pkg/Bom && (cd flat && xar --compression none -cf '../{{.projectName}}-{{.version}}.pkg' * )",
	outputFileExtension:     "pkg",
}
