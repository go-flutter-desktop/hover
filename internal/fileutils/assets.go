package fileutils

import (
	"os"

	rice "github.com/GeertJohan/go.rice"
	"github.com/go-flutter-desktop/hover/internal/log"
)

// AssetsBox hover's assets box
var AssetsBox *rice.Box

func init() {
	var err error
	AssetsBox, err = rice.FindBox("../../assets")
	if err != nil {
		log.Errorf("Failed to find hover assets: %v", err)
		os.Exit(1)
	}
}
