package flutterversion

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/logx"
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
		logx.Errorf("Failed to run %s: %v", logx.Au().Magenta("flutter --version --machine"), err)
		os.Exit(1)
	}

	// Read bytes from the stdout until we receive what looks like the start of
	// a valid json object. This code may be removed when the following flutter
	// issue is resolved. https://github.com/flutter/flutter/issues/54014
	outputBuffer := bytes.NewBuffer(out)
	for {
		b, err := outputBuffer.ReadByte()
		if err != nil {
			logx.Errorf("Failed to run %s: did not return information in json", logx.Au().Magenta("flutter --version --machine"))
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
		logx.Errorf("Failed parsing json: %v", err)
		os.Exit(1)
	}
	return response
}

type flutterVersionResponse struct {
	Channel        string
	EngineRevision string
}
