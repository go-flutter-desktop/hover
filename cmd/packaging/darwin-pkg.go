package packaging

// DarwinPkgPackagingTask packaging for darwin as pkg
var DarwinPkgPackagingTask = &packagingTask{
	packagingFormatName: "darwin-pkg",
	dependsOn: map[*packagingTask]string{
		DarwinBundlePackagingTask: "flat/root/Applications",
	},
	templateFiles: map[string]string{
		"darwin-pkg/PackageInfo.tmpl.tmpl":  "flat/base.pkg/PackageInfo.tmpl",
		"darwin-pkg/Distribution.tmpl.tmpl": "flat/Distribution.tmpl",
	},
	dockerfileContent: []string{
		"FROM ubuntu:bionic",
		"RUN apt-get update && apt-get install cpio git make g++ wget libxml2-dev libssl1.0-dev zlib1g-dev -y",
		"WORKDIR /tmp",
		"RUN git clone https://github.com/hogliux/bomutils && cd bomutils && make > /dev/null && make install > /dev/null",
		"RUN wget https://storage.googleapis.com/google-code-archive-downloads/v2/code.google.com/xar/xar-1.5.2.tar.gz && tar -zxvf xar-1.5.2.tar.gz > /dev/null && cd xar-1.5.2 && ./configure > /dev/null && make > /dev/null && make install > /dev/null",
	},
	packagingScriptTemplate: "(cd flat/root && find . | cpio -o --format odc --owner 0:80 | gzip -c ) > flat/base.pkg/Payload && mkbom -u 0 -g 80 flat/root flat/base.pkg/Bom && (cd flat && xar --compression none -cf '../{{.projectName}}-{{.version}}.pkg' * )",
	outputFileExtension:     "pkg",
}
