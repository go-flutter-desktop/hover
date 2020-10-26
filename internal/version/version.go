package version

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"runtime/debug"
	"sync"

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
	out, err := exec.Command(build.FlutterBin(), "--version", "--machine").Output()
	if err != nil {
		log.Errorf("Failed to run %s: %v", log.Au().Magenta("flutter --version --machine"), err)
		os.Exit(1)
	}

	// Read bytes from the stdout until we receive what looks like the start of
	// a valid json object. This code may be removed when the following flutter
	// issue is resolved. https://github.com/flutter/flutter/issues/54014
	outputBuffer := bytes.NewBuffer(out)
	for {
		b, err := outputBuffer.ReadByte()
		if err != nil {
			log.Errorf("Failed to run %s: did not return information in json", log.Au().Magenta("flutter --version --machine"))
			os.Exit(1)
		}
		if b == '{' {
			outputBuffer.UnreadByte()
			break
		}
	}

	var response flutterVersionResponse
	err = json.NewDecoder(outputBuffer).Decode(&response)
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

var (
	hoverVersionValue string
	hoverVersionOnce  sync.Once
)

func HoverVersion() string {
	hoverVersionOnce.Do(func() {
		buildInfo, ok := debug.ReadBuildInfo()
		if !ok {
			log.Errorf("Cannot obtain version information from hover build. To resolve this, please go-get hover using Go 1.13 or newer.")
			os.Exit(1)
		}
		hoverVersionValue = buildInfo.Main.Version
	})
	return hoverVersionValue
}
