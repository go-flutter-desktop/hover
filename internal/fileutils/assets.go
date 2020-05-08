package fileutils

//go:generate rice embed

import (
	"os"
	"sync"

	rice "github.com/GeertJohan/go.rice"
	"github.com/go-flutter-desktop/hover/internal/logx"
)

var (
	assetsBox     *rice.Box
	assetsBoxOnce sync.Once
)

// AssetsBox hover's assets box
func AssetsBox() *rice.Box {
	assetsBoxOnce.Do(func() {
		var err error
		assetsBox, err = rice.FindBox("../../assets")
		if err != nil {
			logx.Errorf("Failed to find hover assets: %v", err)
			os.Exit(1)
		}
	})
	return assetsBox
}
