package fileutils

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "README.md",
		FileModTime: time.Unix(1587423146, 0),

		Content: string("# assets\n\nThis directory contains templates and config files that hover uses to initialize apps and packaging structures. When modifying these assets, you need to update the generated code so that the assets are included in the Go build process.\n\n## Installing rice\n\nInstall the rice tool by running `(cd $HOME && GO111MODULE=on go get -u -a github.com/GeertJohan/go.rice/rice)`.\n\n## Updating code\n\nRun `go generate ./...` in the repository to update the generated code.\n"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "app/gitignore",
		FileModTime: time.Unix(1571249485, 0),

		Content: string("build\n.last_goflutter_check\n"),
	}
	file5 := &embedded.EmbeddedFile{
		Filename:    "app/go.mod",
		FileModTime: time.Unix(1571249485, 0),

		Content: string(""),
	}
	file6 := &embedded.EmbeddedFile{
		Filename:    "app/hover.yaml.tmpl",
		FileModTime: time.Unix(1587497089, 0),

		Content: string("#application-name: \"{{.applicationName}}\" # Uncomment to modify this value.\n#executable-name: \"{{.executableName}}\" # Uncomment to modify this value. Only lowercase a-z, numbers, underscores and no spaces\n#package-name: \"{{.packageName}}\" # Uncomment to modify this value. Only lowercase a-z, numbers and no underscores or spaces\nlicense: \"\" # MANDATORY: Fill in your SPDX license name: https://spdx.org/licenses\ntarget: lib/main_desktop.dart\nbranch: \"\" # Change to \"@latest\" to download the latest go-flutter version on every build\n# cache-path: \"/home/YOURUSERNAME/.cache/\" #  https://github.com/go-flutter-desktop/go-flutter/issues/184\n# opengl: \"none\" # Uncomment this line if you have trouble with your OpenGL driver (https://github.com/go-flutter-desktop/go-flutter/issues/272)\ndocker: false\nengine-version: \"\" # change to a engine version commit\n"),
	}
	file7 := &embedded.EmbeddedFile{
		Filename:    "app/icon.png",
		FileModTime: time.Unix(1571249485, 0),

		Content: string("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x01\x00\x00\x00\x01\x00\b\x06\x00\x00\x00\\r\xa8f\x00\x00\x0f\x93IDATx\x9c\xed\xddk\x93\x1cU\x19\xc0\xf1\xfd\x14\x99\xe5\xf2\xc6J6\xf1\x82%$\x90\x9d\x0e \x97@Hvg6\xbb\v!\xe4\x86\x10\b\x97\xd2\x0f \xc5E\x04/\x94R(J\xc2\xc6@.@@߈VYR\xa5\\\x04\xa4Pdfvg\x92\r\xa5oԲ\xac\x12- \x90d\x03\x99\xde\xc7\x17\t!\xc9\xce\xf4L\xf7\xf4\xe9s\x9e\xd3\xff\u007f\xd5y\xbfs*\xcfo\xceL\xf7t\xfa\xfahN\xc5)\x19\x1c\xac\xc9ն\xff\x0eJ\xa7`R.\b\xaa\xa1x\xbc\ued7d\xc7\xde\x14L\xcaҠ\x1a~\x14\xd4d\xd4\xf6\xdfB\xe9\x04\x00\xd4U'\xfe\xa1|\x14TC\x01\x00\u007f\x02\x00\xeaX0)\x17\x14+\xe1\a'7\x15\x00\xbc\t\x00(\xb2eSr\xdei\xc3\x0f\x00^\x05\x00Զ\x8bjr~P\rߛ\xb3\xa9\x00\xe0M\x00@-+N\xc9\xe2b%<\xd8rS\x01\xc0\x9b\x00\x80\xe6T\x9c\x92\xc1b5<\xd4vS\x01\xc0\x9b\x00\x80NkpR.)V\xc2#\x91\x9b\n\x00\xde\x04\x00t\xb2\xa5\x15\xb92\xa8\x86G;n*\x00x\x13\x00P____\xdf`M\x86\xba\xdeT\x00\xf0&\x00\xa0\xbe\xa0\"\xd7\x14\xab\xe1\xc7\x00\x90\xbf\x00 \xe7\rN\xca\x15\xb1\x86\x1f\x00\xbc\n\x00r\\\xb1&\x17\a\x95p&\xf6\xa6\x02\x807\x01@N+\xd6\xe4\xe2b%<\x9chS\x01\xc0\x9b\x00 \x87\x9d\x18\xfe\xe8K}\x00\x90\x8b\x00 g\x05U\xb9\xac\xa7\xe1\a\x00\xaf\x02\x80\x1cU\xac\xcaUA7\xd7\xf9\x01 7\x01@N\x1a\xac\xc9P\xb1\x12~\x92ʦ\x02\x807\x01@\x0e\nj2\x1aT\xc2c\xa9m*\x00xӉ_|\xda\x1eRc\xabX\rﱽ\xc7V+V\xa4\x1cT\xc2f\xaa\x1b\v\x00\xde\xc4\t\xc0り\\\x93\xea;?\x00x\x17\x00xZ\xec\xdb{\x01 \x97\x01\x80\x87%\xba\xbd\x17\x00r\x19\x00xV\xb1&\x97\x1b\xdfT\x00\xf0&\x00\xf0\xa8Ԯ\xf3\x03@n\x02\x00OJ\xf5:?\x00\xe4&\x00\xf0\xa0ԯ\xf3\x03@n\x02\x00\xe5\x15'\x9b\xd7g\xbe\xa9\x00\xe0M\x00\xa0\xb8\xc1\x9a\xacN\xfd&\x1f\x00\xc8U\x00\xa0\xb4bUVez\xec\a\x00/\x03\x00\x85\x05UYn\xf4:?\x00\xe4&\x00PV&\xd7\xf9\x01 7\x01\x80\xa22\xbb\xce\x0f\x00\xb9\t\x00\x94\x94\xe9u~\x00\xc8M\x00\xa0\xa0̯\xf3\x03@n\x02\x00ǳv\xa9\x0f\x00r\x11\x008\\P\x95\xe5N\xbd\xf3\x03\x80w\x01\x80\xa3-\xadȥ\x89\xfe\xd3\x0e\x00\xa0\x18\xf1H0\a[Z\x95\xa0\xe7Gw\x03\x00u\x11'\x00\xc7ZZ\x95 \xa8\x86\x1f9\xb0q\x00\x90\x83\x00\xc0\xa1R\xf9O;\x00\x80b\x04\x00\x8eT\xac\xca*\a6\v\x00r\x16\x008\x90S7\xf9\x00@\xae\x02\x00\xcb9w\x93\x0f\x00\xe4*\x00\xb0X\xb1\xd6\xdc\xe0\xc0\x06\x01@\x8e\x03\x00K\rV\x9b\x9b\x8a\xd5pց\r\x02\x80\x1c\a\x00\x16*֚[\x1c\xd8\x18\x00 \x00ȺbE\xc6վ\xf3\x03\x80w\x01@\x86\r\xd6\xe4ju_\xf8\x01\x80\xd7\x01@F-\xadȕ*n\xf2\x01\x80\\\x05\x00\x19\xe4̓|\x00\x80\xce\b\x00LopEJ\x0el\x02\x00P\xcb\x00\xc0`\xea\xee\xf0\x03\x80\xdc\x05\x00\x86\xf2\xe6\v?\x00\xf0:\x000\xd0`U\xbe\xe8\xc0\v\a\x00\xea\x18\x0f\x041PPk\xae\xb7\xfd\xc2\x01\x80\xba\x89\x13\x80\x89M\x05\x00R\x12\x00\x98\xd8T\x00 %\x01\x80\x89M\x05\x00R\x12\x00\x98\xd8T\x00 %\x01\x80\x89M\xad6\xd79\xf0\xc2\x01\x80:\x06\x00&6\x95\x13\x00)\t\x00Ll*\x00\x90\x92\x00\xc0Ħ\xfa\x0e@E\xc62\xdfT2\x92\xef7\x02\x01\x00\x00PD\x00` \x00 -\x01\x80\x81\xb8\n@Z\x02\x00\x03q\x02 -\x01\x80\x81\x00\x80\xb4\x04\x00\x06\x02\x00\xd2\x12\x00\x18\b\x00HK\x00` \x00 -\xf9\x0e\x00\x0f\x04\x01\x00\x8a\xc8w\x00\x02N\x00\x00@\xed\x03\x00\x03\x01\x00i\t\x00\f\x04\x00\xa4%\x000\x10\x00\x90\x96\x00\xc0@\x00@Z\x02\x00\x03\x01\x00i\t\x00\f\x04\x00\xa4%\x000\x10\x00\x90\x96\x00\xc0@\x00@Z\x02\x00\x03\xf1<\x00\xd2\x12\x00\x18\x88\x13\x00i\t\x00\f\x04\x00\xa4%\x000\x10\x00\x90\x96\x00\xc0@\x00@Z\x02\x00\x03\x01\x00i\xc9w\x00x\x1e\x00\x00PD\xbe\x03\x10p\x02\x00\x00j\x1f\x00\x18\b\x00HK\x00` \xef\x01\xe0F o\x02\x00\x03y\x0f\x00'\x00o\x02\x00\x03\x01\x00i\t\x00\f\x04\x00\xa4%\x000\x10\x00\x90\x96\x00\xc0@\x00@Z\x02\x00\x03\x01\x00i\t\x00\f\x04\x00\xa4%\x000\x10\x00\x90\x96\x00\xc0@\x00@Z\x02\x00\x03\xf1H0\xd2\x12\x00\x18\x88\x13\x00i\xc9w\x00\xf890\x00PD\xbe\x03\x10p\x02\x00\x00j\x1f\x00\x18\b\x00HK\x00` \x00 -\x01\x80\x81\x00\x80\xb4\x04\x00\x06\x02\x00\xd2\x12\x00\x18hѯ\xde\xfd\xf9\x97^mJ\x1a\xeb\xbc\xd7\xdc[\x03\xdb\xf7\xffc\xde\xe6\xa7^\xcfr\x15ny\xfa\x8f\xb6\xd7Y\xb7\xba\xb1\xce\xde\xf2Lj뜯\xff\xa2\xf6\xb9{_\x10\x1f\xd7\xfco\xff^\x96\xfd\xee\xdd\x1fg\x0e\xc0\xc0\xf3\xffyq\xd1ˡ\xa4\xb1>\xef\xc0\xfa\xc2\x19k\xd1DC\n\xb7<\x9d\xe9\xea\xbf\xd5\xfe:k\xcb3\xd6\xd7\xd9\x0e\xadsnsw\x9d{\xe7s\xf2\xd5\xdf\xfe[\x86\xeaკ\x03\xb0\xf0\xf9\xff\xfa\r\xc0\xf6}\xb9\x1b~\x00\xd0\x03\xc0\xa7\xc3_n\x84\xb6\x00\xf0\xfc\x04\x90C\x00l\x0f\xbek\x00\xd8\x1e\xf2n\x86\xdf\x1a\x00>}\x048s\xf8m|\x04\xb0=\xfc\xae\x00`{\xe8]\a\xe0\xdc;\x9f\x93\xcb^\xf8l\xf8-\x9e\x00\xfc\xf9\b\xd0\x12\x00N\x00\xb9\x06\xc0\xf6\xa0w\xf3\xce\x0f\x00\x00\x00\x009\x01\xa0\xdd\xf0\xab\xff\b`{\xf8\xf9\b\xc0\xf0k\x1e\xfer#\x94U\xf5\xf0\x01\x00P\f\x80\xed\xe1\a\x00\x87\x01\xb8\xe3\xd9\xc8\xe1\xe7\x04\x00\x00\x00\xe0+\x00w<+\x97\xfe\xe6_\x91ï\xfe;\x00\xdb\xc3\xdf\x16\x80\f\xbf\x03\xb0=\xfc\x00\xa0w\xf89\x01\x18\x18\xfe\xbc\x9d\x00l\x0f\xbe+\xc3\xef\f\x00'\x86\xbfT\xef<\xfc\x00`\b\x80\x85\x8f\xd7e\xde\xe6\xa7R]\x00\x00\x00i\x0f?\x1f\x01L\x010\xd1\x00\x80\x9c\x01\xa0q\xf89\x01\x18\x02`Ѷ))ܼ'\xbd\xd5\x06\x00\xdb\xc3\x0f\x00\x8e\x00p\xca\xf0\xab\x00 \xad\xdf\x02\xb8\n\xc0\xc2mSR\xb8iOz\xcb\xd1w\u007f\x17\x00\xb0=\xf8\xd6\x01\xe8a\xf8U\x9f\x00l\x0f\u007f$\x00\x8fMʼ\x9b\xf6\xa4\xb3nv\xf7\xf8\x0f\x00n\r\xbf\x1a\x00\xd28\x01\xd8\x1e\xfeH\x00\xb6Nɼ\xaf\xedIeq\x05\x00\x00L\r?'\x00S\x00\xfctR\n7\xee\xee}E\x1c\xff\x01 \xc7\x00\xb4\x18~\x00pd\xf8\x8f\x03P\x93\u008d\xbbz\\\xbb\xdb~\xf9\a\x009\x06 \xc5\xe1\xe7#\x80)\x00~2)\x85M\xbb{[7E\x0f?\x00\xb8\x01\x80\vï\x0e\x00\xdfO\x00\x03\x8f\xd6d\xde\xc6]\xc9צ]\x1d\xdf\xfd\x01\xc0\xfe\xf0g\n@\xc4\xf0\xab\x03\xc0\xfb\x13\xc0\xa35)lܕ|u\xf8\xf2\x0f\x00r\x06\x80\xa1\xe1\xe7\x04`\n\x80\x1fU\xa4\xb0ag\xb2\xb5q\xa7\x146\xbb?\xfc\x00\x90\x11\x00\x1d\x86_%\x00\xbe\x9f\x00\x06z\x01\xe0\xc6\xdd*\xde\xfd\x01 \x03\x00\xba\x18~\x95\x00\xf8~\x02\x18x\xe4m\x99\xb7~g\xfc\xb5agW\x9f\xfd\x01\xc0\r\x00\xb4\x0f?\x00\x18\x02`\xe1#oKa\xfd\xce\xf8kSw\xef\xfe.\x00`{\xf8\xbd\x06\xa0\xcb\xe1W\v@\x1a\xbf\x06t\x19\x80\x81\x87\xff\"\x85\x1b\x9e\x8c\xb7\xd6=)\x85\x88\xdb~\x01 '\x00\xc4\x18~\xc5\x00\xf8}\x02\x18x\xf8-)\xac{\"\xdeڸ\xab\xeb\xe1\a\x00O\x01\xc8x\xf8\xf9\b`\n\x80\x1f\xbe%\x85\xb5O\xc4[1\xde\xfd\x01\xc0\xfe\xf0\xa7\x0e@\xcc\xe1\a\x00\x97\x01\xf8AL\x006\xc4{\xf7\a\x00\xfbß*\x00\xb7\xef\x8d=\xfc\xaa\x01\xf0\xfe#@\\\x00:\xfc\xe8\a\x00<\x06\xe0\xf6\xbdr\xe9\xaf\xff\x19{\xf8U\x03\xe0\xfd\t\xe0\xa1?I\xe1\xfa\x1dݭ\xf5;c\x0f?\x000\xfci\x01`\xe5?\x06\xf1\xfd*\xc0\x82\x87\xfe,\x855;\xba[\t\xde\xfd\x01\xc0\x03\x00z\x1c\xfe\\\x9f\x00l\x0f\u007f\xc7\x13\xc0\xf7\xdf\xecn\xf8ox2\xd1\xf0\x03\x80r\x00R\x18~\xd5\x00\xf4\xfa\x1d\x80\xed\xe1\xef\b\xc0\xf7ޔ\xc2u;:\xaf.\u007f\xf4\x03\x00\x1e\x01\x90\xd2\xf0\xab\x06\xc0\xf7\x13\xc0\x82\xef\xbe!\x85k\u007f\x16\xbd\xd6>\x91x\xf8\x01@)\x00\x8e\r?\x00\x18\x05`{\xf4\xea\xf2G?\x00\xe0\t\x00)\x0e\xbfz\x00z\xfd\x12\xd0\xf6\xf0w\x04\xe0;oH\xff\xf8D\xdbUX\xb3\xa3\xa7\xe1\a\x00e\x00\xa4<\xfc\x00\xe0\xc0\x8a\x04\xe0\xc17\xa40\xb6\xbd\xfd\x8a\xf1\xa3\x1f\x00P\x0e\x80\x81\xe1W\x0f\x80\xf7\x1f\x01\x1ex]\nc\x13\xadW\n\xef\xfe\x00\xa0\x04\x00C\xc3\x0f\x00\x0e\xac\x8e\x00\x8cN\xb4^\x9b\xe2\xdf\xf6\v\x00\n\x0108\xfc\xea\xaf\x02,\xf2\x1d\x80\xfb_\x97\xfeՏ\xcf]\xe3\xdbS\x19~\x00p\x1c\x80\f\x86_5\x00\xbe\xdf\n\xbc\xe0\xfeפ\xb0zb\xeeJ\xf0\xa3\x1f\x00P\x06@Fß\x16\x00Vn\x05\xf6\x1e\x80o\xb5\x00`l{W\x0f\xfb\x04\x00\xc5\x00d8\xfc\xaaO\x00\xbe\xff\x1ap\xfe}\xafJ\xa1\xfc\xf8\xe9kC\xb2\x1f\xfd\x00\x80\x9b\x00\xccA \xe3\xe1W\r\x80\xf7'\x80\xfb\xfe \xfd#\xdb>[c\x13\xa9\xbe\xfb\x03\x80\xfd\xe1?\r\x00\v\xc3\x0f\x00\x0e\x030\xff\x9eW\xa4\xbf\xb4\xf5\xe4*\xacO\xfe\xa3\x1f\x00p\x1c\x00K\xc3\x0f\x00\x8e\x03P(m;\xbeF&\xba~\xd47\x00\xe8B\xc0\xe6\xf0\xab\x06\xc0\xf7\xe7\x01̿\xfb\x15)\x94\xb6\x1e_=\xfc\xe4\x17\x00\x1c\x06\xe06\xbbß\x16\x02\x00`\x00\x80\x05w\xbf\"\xfd\xc3[\xa5\xbf\xbc\xcdȻ?\x00X\x06ලr\x89\x03ï\x16\x00\xef?\x02\xdc\xf5\xa2\xf4\x0f=\xd6\xf3O~]\x06\xc0\x05\x04l\x0e\xff\xb0\x03ï\x16\x00\xdfO\x00\xf3\xefzI\n\xc3[c?\xea\x1b\x00\x1c\a\xe0\x94\xe1\a\x00\x00h\x0f\xc07_:\xfe\xc0OC\xc3\x0f\x006\x008}\xf8}A\x80;\x01M\x00p\xf7\xcbF\xdf\xfd]A ?\x00\xec\x95e\xbf\xfc\xfb\x9c\xe1\xf7\x01\x00\xbe\x030\x00\xc0\xc0#o\x1b\x1f~\x00\xc8f\xf8\xcf\xda\xf2L\xd8n\xf8]\x01\xa0\x17\x04\xac\x00\x10Ԛ\xeb\x83j(\xbe\xaebE\xc63\xdfT2\xd2\xf8>\xb9\xa8\xdd\xf0\x03@\xc2\x00\x80\xb4\xd4\t\x00W\x10\x00\x00\x87\x16\x00\xf8\x93\x16\x00\x92\"\x00\x00\x00@\x11\x01\x80\x81\x00\x80\xb4\xa4\t\x80$\b\x00\x00\x00PD\xdd\x00\xe0\x12\x02\x00\xe0\xc0\x02\x00\u007fZs@\x96h\x02 .\x02Vn\x04\x02\x00Ғ\xb6\x13@\\\x008\x01\x00\x00E\xd4-\x00Z\x11\x00\x00\x00\xa0\x88\xb4\x02\xd0-\x02\x00\x00\x00\x14\x11\x00\x18\b\x00HKq\x00Ј\x00\x00\x00\x00E\xa4\x1d\x80N\b\x00\x00\x00PD>\x00\x10\x85\x00\x00\x00\x00E\x14\x17\x00m\b\x00\x00\x00PD\x00` \x00 -%\x01@\x13\x02v\x00\xa86\xd7\xd9\x1eR\xa3\xab\"c\x99o*\x19\xc97\x00\xceD\x80\x13\x00'\x00\x8a()\x00\xae#\xf0)\x04\x00\x00\x00\x14\x91\xcf\x00\x94\xea\x00\x00\x00\x14Y/\x00h@\x00\x00\x00\x80\"\x02\x00\x03\x01\x00i\xa9\xdb\xe7\x01hE\x80\xe7\x01\x00\x00E4Ґ\v}\x06\x80\x13\x00\x00PDi\x00\xe02\x02\x00\x00\x00\x14QZ\x00\xb8\x8a\x00\x00\x18X\x83U\xb96\xf3M%#\x8d\xd6{\xff\x0e\x00\x00\xce\b\x00HKi\x02\xe0\"\x02\x00\x00\x00\x14\xd1\xf8\xb4,N\x13\x00\xd7\x10\xe0*\x00\x00PDi\x9f\x00\\\x03\x80\x13\x00\x00PD&\x00p\t\x01\x00\x00\x00\x8a\xc8\x14\x00\xae \x00\x00\x00@\x11\x99\x04\xc0\x05\x04\x00\x00\x00(\"\x000\x10\x00\x90\x96L\\\x05p\n\x81}R\xce|S\x01\x80\xb4d\xfa\x04`\x11\x81c\xa3\xfb\xe5z+\x9b\n\x00\xa4\xa5\xac\x00\xc8\x18\x81c\xab\x1b\x16\x1f[\a\x00\xa4\xa5,>\x02d\x8d@y\xca\xf2\xbfO\x00 -ey\x020\x8e@#<\xbazZ\x86l\xef)\x00\x90\x9a\xb2>\x01\x18C\xa01;3\xb2_V\xda\xdeϾ\xbe>\x00 =\xd98\x01\x18@\xa092-ö\xf7\xf2d\x00@Z\xb2\t@\x1a\b\f\xd7\xc3١Fs\x93\xed}<-\x00 -\xd9\xfa\b\x90\x1e\x02\xf2\r\xdb{8'\x00 -\xd9>\x01$E`\xa8\x1eή\x98ln\xb6\xbd\u007f-\x03\x00Ғ\v'\x80$\b\\5ټ\xd9\xf6\u07b5\r\x00HK\xae\x9c\x00b@\xd0\\Qon\xb0\xbdo\x91\x01\x00i\xc9E\x00\"\x1086\\\x975\xb6\xf7\xacc\x00@Zr\x15\x80V\bX\xbd\xbd7N\x00@Zr\x19\x80\x93\x104£\xd75\xa4d{\xaf\xba\x0e\x00HK\xee\x030{dlZ\xae\xb6\xbdO\xb1\x02\x00Ғ\xdb\x00\xcc\x1e\x19\u007fG\xae\xb0\xbdG\xb1\v\xaa\xcdu\xb6\x87\xd4\xe8\xaa(\xf9,F\x1ds\x16\x80\xc6\xec\xcc\xe8~Yn{\u007f\x12\xc5\t\x80\xb4\xe4(\x00ǜ\xf9aO\x92\x00\x80\xb4\xe4\x1c\x00SasՔ\\g{_z\n\x00HK\x8e\x01p\xacܰ\xf4\x18\xaf4\x03\x00Ғ+\x00\f\xd5\xc3O\xcau\x19\xb5\xbd\x1f\xa9\x04\x00\xa4\xa5Ѻ,\x19\xb2\r\x80+O\xf2I+\x00 -}\n\x805\x04\x8e\u007fۿ\xca\xf6>\xa4\x1a\x00\x90\x96N\x05\xc0\x02\x02\x87\xcb\xfb\xe5\x12\xdb{\x90z\x00@Z:\x13\x80\xac\x10(\xd5\xc3\x0fǧ\xa5h\xfb\xf5\x1b\t\x00HK6\x00(\xd7\xc3\x0fnxG.\xb4\xfdڍ\x05\x00\xa4\xa5\xf1iY|&\x00&\x11(\xd7\xc3\xf7\xd7\x1c\x90%\xb6_\xb7\xd1\x00\x80\xb4\xd4\xea\x04`\n\x81R=<\xb8vZ\x16\xdb~\xcd\xc6\x03\x00\xd2R\x14\x00)#pht\xbf,\xb5\xfdz3\t\x00HK\x9d\x00H\a\x82\xd9\xc3#\xfb\xe4bۯ5\xb3\x00\x80\xb4\xd4-\x00\x89\x11h\xccάj\xc8\nۯ3\xd3\x00\x80\xb4\x14\a\x80\x04\b\xf8y\x9d\xbfS\x00@Z\x8a\v@\xb7\bx}\x9d\xbfS\x00@ZJ\x02@'\x04\xbc\xbf\xce\xdf)\x00 -\xb5\xbb\x0f )\x02\xe5\xc6\xec\xff\xd6\xfb~\x9d\xbfS<\x12\x8c\xb4\x94\xf4\x04\xd0\n\x81R#|o\xcd\x01\xf9\xb2\xed\xd7d=N\x00\xa4\xa5^\x018\x05\x82\x83\xa3\u007f\x95\xf3m\xbf\x1e'\x02\x00\xd2R\x1a\x00\f\xd7Ã\xb9\xb9ɧ\x9b\x00\x80\xb4\x94\xc2G\x80Ck\xf3\xfam\u007f\xbb\x00\x80\xb4\xd4\v\x00\xc3\xf5\xd9#c\a\xe42ۯ\xc1\xb9\x00\x80\xb4\x94\xfc2\xe0\xeca\x95\xffiG\x16\x01\x00i)!\x003\xa5\x86\\n\xfbow6\x00 -ž\x15\xb8\x11άh(\xfb\xbf\xfa\xb2\x0e\x00HK1\u007f\f\xf4q\xf9@\xce~ؓ$\x00 -\xc5\x01\xa0\\\x97ն\xff^\x15\x01\x00i\xa9\x1b\x00\x86\x1b\xe1ѕ\x93r\x8d\xed\xbfUM\x00@Z\xea\b\xc0\xd4쑕|\xe1\x17/\x00 -u\x00\xe0\xc3Uyz\x92OZ\x01\x00i\xa9\x1d\x00\xc3\xf5\xf0\x83\x91\xbf\xc9\x05\xb6\xff>\x95\x01\x00i\xa9\x15\x00\xa5z\xf8\xfeH#ǿ\xe7\xef5\x00 -\xb5\x00\xe0\x90\xf7\xcf\xed7\x1d\x00\x90\x96N\xff\xcfAg\x0f\xf3\x99?\x85\x00\x80\xb4t\xf2\xbf\ao\x843\xa5)Yn\xfb\xef\xf1\"\x00 -\x8d\xd6e\xc9p=\xfcxhJV\xda\xfe[\xbc\t\x00HK\xa5}\xf2\x95\x95\r\x1e\xf1֪\xff\x03ɉ>)\x8dx\xe1\xbb\x00\x00\x00\x00IEND\xaeB`\x82"),
	}
	file8 := &embedded.EmbeddedFile{
		Filename:    "app/main.go",
		FileModTime: time.Unix(1571249485, 0),

		Content: string("package main\n\nimport (\n\t\"fmt\"\n\t\"image\"\n\t_ \"image/png\"\n\t\"os\"\n\t\"path/filepath\"\n\t\"strings\"\n\n\t\"github.com/go-flutter-desktop/go-flutter\"\n\t\"github.com/pkg/errors\"\n)\n\n// vmArguments may be set by hover at compile-time\nvar vmArguments string\n\nfunc main() {\n\t// DO NOT EDIT, add options in options.go\n\tmainOptions := []flutter.Option{\n\t\tflutter.OptionVMArguments(strings.Split(vmArguments, \";\")),\n\t\tflutter.WindowIcon(iconProvider),\n\t}\n\terr := flutter.Run(append(options, mainOptions...)...)\n\tif err != nil {\n\t\tfmt.Println(err)\n\t\tos.Exit(1)\n\t}\n}\n\nfunc iconProvider() ([]image.Image, error) {\n\texecPath, err := os.Executable()\n\tif err != nil {\n\t\treturn nil, errors.Wrap(err, \"failed to resolve executable path\")\n\t}\n\texecPath, err = filepath.EvalSymlinks(execPath)\n\tif err != nil {\n\t\treturn nil, errors.Wrap(err, \"failed to eval symlinks for executable path\")\n\t}\n\timgFile, err := os.Open(filepath.Join(filepath.Dir(execPath), \"assets\", \"icon.png\"))\n\tif err != nil {\n\t\treturn nil, errors.Wrap(err, \"failed to open assets/icon.png\")\n\t}\n\timg, _, err := image.Decode(imgFile)\n\tif err != nil {\n\t\treturn nil, errors.Wrap(err, \"failed to decode image\")\n\t}\n\treturn []image.Image{img}, nil\n}\n"),
	}
	file9 := &embedded.EmbeddedFile{
		Filename:    "app/main_desktop.dart",
		FileModTime: time.Unix(1587299806, 0),

		Content: string("import 'package:flutter/foundation.dart'\n    show debugDefaultTargetPlatformOverride;\nimport 'package:flutter/material.dart';\n\nimport 'main.dart' as original_main;\n\nvoid main() {\n  debugDefaultTargetPlatformOverride = TargetPlatform.fuchsia;\n  original_main.main();\n}\n"),
	}
	filea := &embedded.EmbeddedFile{
		Filename:    "app/options.go",
		FileModTime: time.Unix(1571249485, 0),

		Content: string("package main\n\nimport (\n\t\"github.com/go-flutter-desktop/go-flutter\"\n)\n\nvar options = []flutter.Option{\n\tflutter.WindowInitialDimensions(800, 1280),\n}\n"),
	}
	filec := &embedded.EmbeddedFile{
		Filename:    "packaging/README.md",
		FileModTime: time.Unix(1587470036, 0),

		Content: string("# packaging\nThe template files in the subdirectories are only copied on init and then executed on build.\n"),
	}
	filee := &embedded.EmbeddedFile{
		Filename:    "packaging/darwin-bundle/Info.plist.tmpl",
		FileModTime: time.Unix(1587472853, 0),

		Content: string("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE plist PUBLIC \"-//Apple Computer//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n<plist version=\"1.0\">\n    <dict>\n        <key>CFBundleDevelopmentRegion</key>\n        <string>English</string>\n        <key>CFBundleExecutable</key>\n        <string>{{.executableName}}</string>\n        <key>CFBundleGetInfoString</key>\n        <string>{{.description}}</string>\n        <key>CFBundleIconFile</key>\n        <string>icon.icns</string>\n        <key>CFBundleIdentifier</key>\n        <string>{{.organizationName}}</string>\n        <key>CFBundleInfoDictionaryVersion</key>\n        <string>6.0</string>\n        <key>CFBundleLongVersionString</key>\n        <string>{{.version}}</string>\n        <key>CFBundleName</key>\n        <string>{{.applicationName}}</string>\n        <key>CFBundlePackageType</key>\n        <string>APPL</string>\n        <key>CFBundleShortVersionString</key>\n        <string>{{.version}}</string>\n        <key>CFBundleSignature</key>\n        <string>{{.organizationName}}.{{.packageName}}</string>\n        <key>CFBundleVersion</key>\n        <string>{{.version}}</string>\n        <key>CSResourcesFileMapped</key>\n        <true/>\n        <key>NSHumanReadableCopyright</key>\n        <string></string>\n    </dict>\n</plist>"),
	}
	fileg := &embedded.EmbeddedFile{
		Filename:    "packaging/darwin-pkg/Distribution.tmpl",
		FileModTime: time.Unix(1587472689, 0),

		Content: string("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<installer-gui-script minSpecVersion=\"1\">\n\t<title>{{.applicationName}}</title>\n\t<background alignment=\"topleft\" file=\"root/Applications/{{.applicationName}} {{.version}}.app/Contents/MacOS/assets/icon.png\"/>\n\t<choices-outline>\n\t    <line choice=\"choiceBase\"/>\n    </choices-outline>\n    <choice id=\"choiceBase\" title=\"base\">\n        <pkg-ref id=\"{{.organizationName}}.base.pkg\"/>\n    </choice>\n    <pkg-ref id=\"{{.organizationName}}.base.pkg\" version=\"{{.version}}\" auth=\"Root\">#base.pkg</pkg-ref>\n</installer-gui-script>\n"),
	}
	fileh := &embedded.EmbeddedFile{
		Filename:    "packaging/darwin-pkg/PackageInfo.tmpl",
		FileModTime: time.Unix(1587473491, 0),

		Content: string("<pkg-info format-version=\"2\" identifier=\"{{.organizationName}}.base.pkg\" version=\"{{.version}}\" install-location=\"/\" auth=\"root\">\n\t<bundle-version>\n\t\t<bundle id=\"{{.organizationName}}\" CFBundleIdentifier=\"{{.organizationName}}.{{.packageName}}\" path=\"./Applications/{{.applicationName}} {{.version}}.app\" CFBundleVersion=\"{{.version}}\"/>\n    </bundle-version>\n</pkg-info>"),
	}
	filej := &embedded.EmbeddedFile{
		Filename:    "packaging/linux/app.desktop.tmpl",
		FileModTime: time.Unix(1587470111, 0),

		Content: string("[Desktop Entry]\nVersion=1.0\nType=Application\nTerminal=false\nCategories=\nName={{.applicationName}}\nIcon={{.iconPath}}\nExec={{.executablePath}}"),
	}
	filek := &embedded.EmbeddedFile{
		Filename:    "packaging/linux/bin.tmpl",
		FileModTime: time.Unix(1587423157, 0),

		Content: string("#!/bin/sh\n/usr/lib/{{.packageName}}/{{.executableName}}"),
	}
	filem := &embedded.EmbeddedFile{
		Filename:    "packaging/linux-appimage/AppRun.tmpl",
		FileModTime: time.Unix(1587423157, 0),

		Content: string("#!/bin/sh\ncd \"$(dirname \"$0\")\"\nexec ./build/{{.executableName}}"),
	}
	fileo := &embedded.EmbeddedFile{
		Filename:    "packaging/linux-deb/control.tmpl",
		FileModTime: time.Unix(1587423157, 0),

		Content: string("Package: {{.packageName}}\nArchitecture: amd64\nMaintainer: @{{.author}}\nPriority: optional\nVersion: {{.version}}\nDescription: {{.description}}\n"),
	}
	fileq := &embedded.EmbeddedFile{
		Filename:    "packaging/linux-pkg/PKGBUILD.tmpl",
		FileModTime: time.Unix(1587471688, 0),

		Content: string("pkgname={{.packageName}}\npkgver={{.version}}\npkgrel={{.release}}\npkgdesc=\"{{.description}}\"\narch=(\"x86_64\")\nlicense=('{{.license}}')\n\npackage() {\n    mkdir -p $pkgdir/\n    cp * $pkgdir/ -r\n}\n"),
	}
	files := &embedded.EmbeddedFile{
		Filename:    "packaging/linux-rpm/app.spec.tmpl",
		FileModTime: time.Unix(1587471688, 0),

		Content: string("Name: {{.packageName}}\nVersion: {{.version}}\nRelease: {{.release}}\nSummary: {{.description}}\nLicense: {{.license}}\n\n%description\n{{.description}}\n\n%install\nmkdir -p $RPM_BUILD_ROOT%{_bindir}\nmkdir -p $RPM_BUILD_ROOT/usr/lib/{{.packageName}}\nmkdir -p $RPM_BUILD_ROOT%{_datadir}/applications\ncp -R $RPM_BUILD_DIR/{{.packageName}}-{{.version}}-{{.release}}.x86_64/* $RPM_BUILD_ROOT\nchmod 0755 $RPM_BUILD_ROOT%{_bindir}/{{.executableName}}\nchmod 0755 $RPM_BUILD_ROOT%{_datadir}/applications/{{.executableName}}.desktop\n\n%files\n%{_bindir}/{{.executableName}}\n/usr/lib/{{.packageName}}/\n%{_datadir}/applications/{{.executableName}}.desktop"),
	}
	fileu := &embedded.EmbeddedFile{
		Filename:    "packaging/linux-snap/snapcraft.yaml.tmpl",
		FileModTime: time.Unix(1587423157, 0),

		Content: string("name: {{.packageName}}\nbase: core18\nversion: '{{.version}}'\nsummary: {{.description}}\ndescription: |\n  {{.description}}\nconfinement: devmode\ngrade: devel\napps:\n  {{.packageName}}:\n    command: {{.executableName}}\n    desktop: local/{{.executableName}}.desktop\nparts:\n  desktop:\n    plugin: dump\n    source: snap\n  assets:\n    plugin: dump\n    source: build/assets\n  app:\n    plugin: dump\n    source: build\n    stage-packages:\n      - libx11-6\n      - libxrandr2\n      - libxcursor1\n      - libxinerama1\n"),
	}
	filew := &embedded.EmbeddedFile{
		Filename:    "packaging/windows-msi/app.wxs.tmpl",
		FileModTime: time.Unix(1587428338, 0),

		Content: string("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Wix xmlns=\"http://schemas.microsoft.com/wix/2006/wi\">\n    <Product Id=\"*\" UpgradeCode=\"*\" Version=\"{{.version}}\" Language=\"1033\" Name=\"{{.applicationName}}\" Manufacturer=\"{{.author}}\">\n        <Package InstallerVersion=\"300\" Compressed=\"yes\"/>\n        <Media Id=\"1\" Cabinet=\"{{.packageName}}.cab\" EmbedCab=\"yes\" />\n        <Directory Id=\"TARGETDIR\" Name=\"SourceDir\">\n            <Directory Id=\"ProgramFilesFolder\">\n                <Directory Id=\"APPLICATIONROOTDIRECTORY\" Name=\"{{.applicationName}}\">\n                    <Directory Id=\"ASSETSDIRECTORY\" Name=\"assets\"/>\n                    <Directory Id=\"FLUTTERASSETSDIRECTORY\" Name=\"flutter_assets\">\n                        <?include directories.wxi ?>\n                    </Directory>\n                </Directory>\n            </Directory>\n            <Directory Id=\"ProgramMenuFolder\">\n                <Directory Id=\"ApplicationProgramsFolder\" Name=\"{{.applicationName}}\"/>\n            </Directory>\n        </Directory>\n        <Icon Id=\"ShortcutIcon\" SourceFile=\"build/assets/icon.ico\"/>\n        <DirectoryRef Id=\"APPLICATIONROOTDIRECTORY\">\n            <Component Id=\"{{.executableName}}.exe\" Guid=\"*\">\n                <File Id=\"{{.executableName}}.exe\" Source=\"build/{{.executableName}}.exe\" KeyPath=\"yes\"/>\n            </Component>\n            <Component Id=\"flutter_engine.dll\" Guid=\"*\">\n                <File Id=\"flutter_engine.dll\" Source=\"build/flutter_engine.dll\" KeyPath=\"yes\"/>\n            </Component>\n            <Component Id=\"icudtl.dat\" Guid=\"*\">\n                <File Id=\"icudtl.dat\" Source=\"build/icudtl.dat\" KeyPath=\"yes\"/>\n            </Component>\n        </DirectoryRef>\n        <DirectoryRef Id=\"ASSETSDIRECTORY\">\n            <Component Id=\"icon.png\" Guid=\"*\">\n                <File Id=\"icon.png\" Source=\"build/assets/icon.png\" KeyPath=\"yes\"/>\n            </Component>\n        </DirectoryRef>\n        <?include directory_refs.wxi ?>\n        <DirectoryRef Id=\"ApplicationProgramsFolder\">\n            <Component Id=\"ApplicationShortcut\" Guid=\"*\">\n                <Shortcut Id=\"ApplicationStartMenuShortcut\"\n                          Name=\"{{.applicationName}}\"\n                          Description=\"{{.description}}\"\n                          Target=\"[#{{.executableName}}.exe]\"\n                          WorkingDirectory=\"APPLICATIONROOTDIRECTORY\"\n                          Icon=\"ShortcutIcon\"/>\n                <RemoveFolder Id=\"CleanUpShortCut\" On=\"uninstall\"/>\n                <RegistryValue Root=\"HKCU\" Key=\"Software\\{{.author}}\\{{.packageName}}\" Name=\"installed\" Type=\"integer\" Value=\"1\" KeyPath=\"yes\"/>\n            </Component>\n        </DirectoryRef>\n        <Feature Id=\"MainApplication\" Title=\"{{.applicationName}}\" Level=\"1\">\n            <ComponentRef Id=\"{{.executableName}}.exe\"/>\n            <ComponentRef Id=\"flutter_engine.dll\"/>\n            <ComponentRef Id=\"icudtl.dat\"/>\n            <ComponentRef Id=\"icon.png\"/>\n            <ComponentRef Id=\"ApplicationShortcut\"/>\n            <?include component_refs.wxi ?>\n        </Feature>\n    </Product>\n</Wix>"),
	}
	filey := &embedded.EmbeddedFile{
		Filename:    "plugin/README.md.dlib.tmpl",
		FileModTime: time.Unix(1587299806, 0),

		Content: string("The `dlib` folder is used for the plugins which use `cgo`.\n\nIf your go-flutter plugin dose't use `cgo`, just ignore this file and the `dlib` folder.\n\nWhen you need to link prebuild dynamic libraries and frameworks,\nyou should copy the prebuild dynamic libraries and frameworks to `dlib`/${os} folder.\n\n`hover plugins get` copy this files to path `./go/build/intermediates` of go-flutter app project.\n`hover run` copy files from `./go/build/intermediates/${targetOS}` to `./go/build/outputs/${targetOS}`.\nAnd `-L{./go/build/outputs/${targetOS}}` is appended to `cgoLdflags` automatically.\nAlso `-F{./go/build/outputs/${targetOS}}` is appended to `cgoLdflags` on Mac OS\n\nAttention: `hover` can't resolve the conflicts\nif two different go-flutter plugins have file with the same name in there dlib folder\n"),
	}
	filez := &embedded.EmbeddedFile{
		Filename:    "plugin/README.md.tmpl",
		FileModTime: time.Unix(1577653312, 0),

		Content: string("# {{.pluginName}}\n\nThis Go package implements the host-side of the Flutter [{{.pluginName}}](https://{{.urlVSCRepo}}) plugin.\n\n## Usage\n\nImport as:\n\n```go\nimport {{.pluginName}} \"{{.urlVSCRepo}}/go\"\n```\n\nThen add the following option to your go-flutter [application options](https://github.com/go-flutter-desktop/go-flutter/wiki/Plugin-info):\n\n```go\nflutter.AddPlugin(&{{.pluginName}}.{{.structName}}{}),\n```\n"),
	}
	file10 := &embedded.EmbeddedFile{
		Filename:    "plugin/import.go.tmpl.tmpl",
		FileModTime: time.Unix(1577653312, 0),

		Content: string("package main\n\n// DO NOT EDIT, this file is generated by hover at compile-time for the {{.pluginName}} plugin.\n\nimport (\n\tflutter \"github.com/go-flutter-desktop/go-flutter\"\n\t{{.pluginName}} \"{{.urlVSCRepo}}/go\"\n)\n\nfunc init() {\n\t// Only the init function can be tweaked by plugin maker.\n\toptions = append(options, flutter.AddPlugin(&{{.pluginName}}.{{.structName}}{}))\n}\n"),
	}
	file11 := &embedded.EmbeddedFile{
		Filename:    "plugin/plugin.go.tmpl",
		FileModTime: time.Unix(1577653312, 0),

		Content: string("package {{.pluginName}}\n\nimport (\n\tflutter \"github.com/go-flutter-desktop/go-flutter\"\n\t\"github.com/go-flutter-desktop/go-flutter/plugin\"\n)\n\nconst channelName = \"{{.pluginName}}\"\n\n// {{.structName}} implements flutter.Plugin and handles method.\ntype {{.structName}} struct{}\n\nvar _ flutter.Plugin = &{{.structName}}{} // compile-time type check\n\n// InitPlugin initializes the plugin.\nfunc (p *{{.structName}}) InitPlugin(messenger plugin.BinaryMessenger) error {\n\tchannel := plugin.NewMethodChannel(messenger, channelName, plugin.StandardMethodCodec{})\n\tchannel.HandleFunc(\"getPlatformVersion\", p.handlePlatformVersion)\n\treturn nil\n}\n\nfunc (p *{{.structName}}) handlePlatformVersion(arguments interface{}) (reply interface{}, err error) {\n\treturn \"go-flutter \" + flutter.PlatformVersion, nil\n}\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1587423146, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "README.md"

		},
	}
	dir3 := &embedded.EmbeddedDir{
		Filename:   "app",
		DirModTime: time.Unix(1587497089, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file4, // "app/gitignore"
			file5, // "app/go.mod"
			file6, // "app/hover.yaml.tmpl"
			file7, // "app/icon.png"
			file8, // "app/main.go"
			file9, // "app/main_desktop.dart"
			filea, // "app/options.go"

		},
	}
	dirb := &embedded.EmbeddedDir{
		Filename:   "packaging",
		DirModTime: time.Unix(1587470036, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			filec, // "packaging/README.md"

		},
	}
	dird := &embedded.EmbeddedDir{
		Filename:   "packaging/darwin-bundle",
		DirModTime: time.Unix(1587472853, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			filee, // "packaging/darwin-bundle/Info.plist.tmpl"

		},
	}
	dirf := &embedded.EmbeddedDir{
		Filename:   "packaging/darwin-pkg",
		DirModTime: time.Unix(1587473491, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			fileg, // "packaging/darwin-pkg/Distribution.tmpl"
			fileh, // "packaging/darwin-pkg/PackageInfo.tmpl"

		},
	}
	diri := &embedded.EmbeddedDir{
		Filename:   "packaging/linux",
		DirModTime: time.Unix(1587470111, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			filej, // "packaging/linux/app.desktop.tmpl"
			filek, // "packaging/linux/bin.tmpl"

		},
	}
	dirl := &embedded.EmbeddedDir{
		Filename:   "packaging/linux-appimage",
		DirModTime: time.Unix(1587423157, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			filem, // "packaging/linux-appimage/AppRun.tmpl"

		},
	}
	dirn := &embedded.EmbeddedDir{
		Filename:   "packaging/linux-deb",
		DirModTime: time.Unix(1587423157, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			fileo, // "packaging/linux-deb/control.tmpl"

		},
	}
	dirp := &embedded.EmbeddedDir{
		Filename:   "packaging/linux-pkg",
		DirModTime: time.Unix(1587471688, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			fileq, // "packaging/linux-pkg/PKGBUILD.tmpl"

		},
	}
	dirr := &embedded.EmbeddedDir{
		Filename:   "packaging/linux-rpm",
		DirModTime: time.Unix(1587471688, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			files, // "packaging/linux-rpm/app.spec.tmpl"

		},
	}
	dirt := &embedded.EmbeddedDir{
		Filename:   "packaging/linux-snap",
		DirModTime: time.Unix(1587423157, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			fileu, // "packaging/linux-snap/snapcraft.yaml.tmpl"

		},
	}
	dirv := &embedded.EmbeddedDir{
		Filename:   "packaging/windows-msi",
		DirModTime: time.Unix(1587428338, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			filew, // "packaging/windows-msi/app.wxs.tmpl"

		},
	}
	dirx := &embedded.EmbeddedDir{
		Filename:   "plugin",
		DirModTime: time.Unix(1587299806, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			filey,  // "plugin/README.md.dlib.tmpl"
			filez,  // "plugin/README.md.tmpl"
			file10, // "plugin/import.go.tmpl.tmpl"
			file11, // "plugin/plugin.go.tmpl"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{
		dir3, // "app"
		dirb, // "packaging"
		dirx, // "plugin"

	}
	dir3.ChildDirs = []*embedded.EmbeddedDir{}
	dirb.ChildDirs = []*embedded.EmbeddedDir{
		dird, // "packaging/darwin-bundle"
		dirf, // "packaging/darwin-pkg"
		diri, // "packaging/linux"
		dirl, // "packaging/linux-appimage"
		dirn, // "packaging/linux-deb"
		dirp, // "packaging/linux-pkg"
		dirr, // "packaging/linux-rpm"
		dirt, // "packaging/linux-snap"
		dirv, // "packaging/windows-msi"

	}
	dird.ChildDirs = []*embedded.EmbeddedDir{}
	dirf.ChildDirs = []*embedded.EmbeddedDir{}
	diri.ChildDirs = []*embedded.EmbeddedDir{}
	dirl.ChildDirs = []*embedded.EmbeddedDir{}
	dirn.ChildDirs = []*embedded.EmbeddedDir{}
	dirp.ChildDirs = []*embedded.EmbeddedDir{}
	dirr.ChildDirs = []*embedded.EmbeddedDir{}
	dirt.ChildDirs = []*embedded.EmbeddedDir{}
	dirv.ChildDirs = []*embedded.EmbeddedDir{}
	dirx.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`../../assets`, &embedded.EmbeddedBox{
		Name: `../../assets`,
		Time: time.Unix(1587423146, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"":                         dir1,
			"app":                      dir3,
			"packaging":                dirb,
			"packaging/darwin-bundle":  dird,
			"packaging/darwin-pkg":     dirf,
			"packaging/linux":          diri,
			"packaging/linux-appimage": dirl,
			"packaging/linux-deb":      dirn,
			"packaging/linux-pkg":      dirp,
			"packaging/linux-rpm":      dirr,
			"packaging/linux-snap":     dirt,
			"packaging/windows-msi":    dirv,
			"plugin":                   dirx,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"README.md":                                file2,
			"app/gitignore":                            file4,
			"app/go.mod":                               file5,
			"app/hover.yaml.tmpl":                      file6,
			"app/icon.png":                             file7,
			"app/main.go":                              file8,
			"app/main_desktop.dart":                    file9,
			"app/options.go":                           filea,
			"packaging/README.md":                      filec,
			"packaging/darwin-bundle/Info.plist.tmpl":  filee,
			"packaging/darwin-pkg/Distribution.tmpl":   fileg,
			"packaging/darwin-pkg/PackageInfo.tmpl":    fileh,
			"packaging/linux/app.desktop.tmpl":         filej,
			"packaging/linux/bin.tmpl":                 filek,
			"packaging/linux-appimage/AppRun.tmpl":     filem,
			"packaging/linux-deb/control.tmpl":         fileo,
			"packaging/linux-pkg/PKGBUILD.tmpl":        fileq,
			"packaging/linux-rpm/app.spec.tmpl":        files,
			"packaging/linux-snap/snapcraft.yaml.tmpl": fileu,
			"packaging/windows-msi/app.wxs.tmpl":       filew,
			"plugin/README.md.dlib.tmpl":               filey,
			"plugin/README.md.tmpl":                    filez,
			"plugin/import.go.tmpl.tmpl":               file10,
			"plugin/plugin.go.tmpl":                    file11,
		},
	})
}
