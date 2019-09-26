package packaging

import (
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
	"os"
)

var linuxPackagingDependencies = []string{"libx11-6", "libxrandr2", "libxcursor1", "libxinerama1", "libglu1-mesa", "libgcc1", "libstdc++6", "libtinfo5", "zlib1g"}

func createLinuxDesktopFile(desktopFilePath string, packagingFormat string, exec string, icon string) {
	desktopFile, err := os.Create(desktopFilePath)
	if err != nil {
		log.Errorf("Failed to create %s.desktop %s: %v", pubspec.GetPubSpec().Name, desktopFilePath, err)
		os.Exit(1)
	}
	desktopFileContent := []string{
		"[Desktop Entry]",
		"Encoding=UTF-8",
		"Version=" + pubspec.GetPubSpec().Version,
		"Type=Application",
		"Terminal=false",
		"Exec=" + exec,
		"Name=" + pubspec.GetPubSpec().Name,
		"Icon=" + icon,
	}

	for _, line := range desktopFileContent {
		if _, err := desktopFile.WriteString(line + "\n"); err != nil {
			log.Errorf("Could not write %s.desktop: %v", pubspec.GetPubSpec().Name, err)
			os.Exit(1)
		}
	}
	err = desktopFile.Close()
	if err != nil {
		log.Errorf("Could not close %s.desktop: %v", pubspec.GetPubSpec().Name, err)
		os.Exit(1)
	}
}
