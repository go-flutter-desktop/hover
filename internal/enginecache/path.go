package enginecache

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-flutter-desktop/hover/internal/log"
	homedir "github.com/mitchellh/go-homedir"
)

func cachePath() string {
	homePath, err := homedir.Dir()
	if err != nil {
		log.Errorf("Failed to resolve home path: %v", err)
		os.Exit(1)
	}

	var p string
	switch runtime.GOOS {
	case "linux":
		p = filepath.Join(homePath, ".cache")
	case "darwin":
		p = filepath.Join(homePath, "Library", "Caches")
	case "windows":
		p = filepath.Join(homePath, "AppData", "Local")
	default:
		log.Errorf("Cannot run on %s, enginecache not implemented.", runtime.GOOS)
		os.Exit(1)
	}
	return p
}
