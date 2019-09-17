package enginecache

import (
	"github.com/logrusorgru/aurora"
	"os"
	"os/exec"
	"regexp"

	"github.com/go-flutter-desktop/hover/internal/log"
)

func flutterRequiredEngineVersion() string {
	out, err := exec.Command("flutter", "--version").Output()
	if err != nil {
		log.Fatal("Failed to run `%s`: %v", aurora.Magenta("flutter --version"), err)
		os.Exit(1)
	}

	regexpEngineVersion := regexp.MustCompile(`Engine • revision (\w{10})`)
	versionMatch := regexpEngineVersion.FindStringSubmatch(string(out))
	if len(versionMatch) != 2 {
		log.Fatal("Failed to obtain engine version")
		os.Exit(1)
	}

	return versionMatch[1]
}
