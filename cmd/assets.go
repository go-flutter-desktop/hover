package cmd

import (
	"fmt"
	"os"

	rice "github.com/GeertJohan/go.rice"
)

var assetsBox *rice.Box

func init() {
	var err error
	assetsBox, err = rice.FindBox("../assets")
	if err != nil {
		fmt.Printf("Failed to find glutter assets: %v", err)
		os.Exit(1)
	}
}
