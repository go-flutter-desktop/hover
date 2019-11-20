package enginecache

import (
	"os"
	"os/exec"
	"regexp"

	"github.com/go-flutter-desktop/hover/internal/log"
)

// FlutterRequiredEngineVersion retunrs the commit id of the engine in use
func FlutterRequiredEngineVersion() string {
	out, err := exec.Command("flutter", "--version").Output()
	if err != nil {
		log.Errorf("Failed to run %s: %v", log.Au().Magenta("flutter --version"), err)
		os.Exit(1)
	}

	regexpEngineVersion := regexp.MustCompile(`Engine • revision (\w{10})`)
	versionMatch := regexpEngineVersion.FindStringSubmatch(string(out))
	if len(versionMatch) != 2 {
		log.Errorf("Failed to obtain engine version")
		os.Exit(1)
	}

	return versionMatch[1]
}
