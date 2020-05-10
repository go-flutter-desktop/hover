package enginecache

import (
	"os"

	"github.com/go-flutter-desktop/hover/internal/log"
)

// DefaultCachePath tries to resolve the user cache directory. DefaultCachePath
// may return an empty string when none was found, in that case it will print a
// warning to the user.
func DefaultCachePath() string {
	cachePath, err := os.UserCacheDir()
	if err != nil {
		log.Warnf("Failed to resolve cache path: %v", err)
	}
	return cachePath
}
