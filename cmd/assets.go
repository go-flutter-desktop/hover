package cmd

import (
	"os"

	rice "github.com/GeertJohan/go.rice"
	"github.com/go-flutter-desktop/hover/internal/log"
)

var assetsBox *rice.Box

func init() {
	var err error
	assetsBox, err = rice.FindBox("../assets")
	if err != nil {
		log.Errorf("Failed to find hover assets: %v", err)
		os.Exit(1)
	}
}
