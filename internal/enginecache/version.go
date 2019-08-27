package enginecache

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

func flutterRequiredEngineVersion() string {
	out, err := exec.Command("flutter", "--version").Output()
	if err != nil {
		fmt.Printf("hover: Failed to run `flutter --version`: %v\n", err)
		os.Exit(1)
	}

	regexpEngineVersion := regexp.MustCompile(`Engine • revision (\w{10})`)
	versionMatch := regexpEngineVersion.FindStringSubmatch(string(out))
	if len(versionMatch) != 2 {
		fmt.Printf("hover: Failed to obtain engine version")
		os.Exit(1)
	}

	return versionMatch[1]
}
