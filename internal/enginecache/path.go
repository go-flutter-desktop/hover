package enginecache

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-flutter-desktop/hover/internal/logx"
)

// DefaultCachePath returns the default root directory for caching data.
func DefaultCachePath() string {
	// TODO: change to os.UserCacheDir()?
	homePath, err := os.UserHomeDir()
	if err != nil {
		logx.Errorf("Failed to resolve home path: %v", err)
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
		logx.Errorf("Cannot run on %s, enginecache not implemented.", runtime.GOOS)
		os.Exit(1)
	}
	return p
}
