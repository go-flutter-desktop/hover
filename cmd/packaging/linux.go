package packaging

import (
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

var linuxPackagingDependencies = []string{"libx11-6", "libxrandr2", "libxcursor1", "libxinerama1", "libglu1-mesa", "libgcc1", "libstdc++6", "libtinfo5", "zlib1g"}

func createLinuxDesktopFile(desktopFilePath string, exec string, icon string) {
	templateData := map[string]string{
		"projectName": pubspec.GetPubSpec().Name,
		"icon":        icon,
		"exec":        exec,
	}

	fileutils.CopyTemplateFromAssetsBox("packaging/app.desktop.tmpl", desktopFilePath, fileutils.AssetsBox, templateData)
}
