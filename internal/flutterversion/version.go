package flutterversion

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
)

// FlutterRequiredEngineVersion returns the commit hash of the engine in use
func FlutterRequiredEngineVersion() string {
	return readFlutterVersion().EngineRevision
}

// FlutterChannel returns the channel of the flutter installation
func FlutterChannel() string {
	return readFlutterVersion().Channel
}

func readFlutterVersion() flutterVersionResponse {
	out, err := exec.Command(build.FlutterBin, "--version", "--machine").Output()
	if err != nil {
		log.Errorf("Failed to run %s: %v", log.Au().Magenta("flutter --version --machine"), err)
		os.Exit(1)
	}
	var response flutterVersionResponse
	err = json.Unmarshal(out, &response)
	if err != nil {
		log.Errorf("Failed parsing json: %v", err)
		os.Exit(1)
	}
	return response
}

type flutterVersionResponse struct {
	Channel        string
	EngineRevision string
}
