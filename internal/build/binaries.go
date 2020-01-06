package build

import (
	"os"
	"os/exec"
	"sync"

	"github.com/go-flutter-desktop/hover/internal/log"
)

var (
	goBin     string
	goBinOnce sync.Once
)

func GoBin() string {
	goBinOnce.Do(func() {
		var err error
		goBin, err = exec.LookPath("go")
		if err != nil {
			log.Errorf("Failed to lookup `go` executable: %s. Please install go or add `--docker` to run the Hover command in a Docker container.\nhttps://golang.org/doc/install", err)
			os.Exit(1)
		}
	})
	return goBin
}

var (
	flutterBin     string
	flutterBinOnce sync.Once
)

func FlutterBin() string {
	flutterBinOnce.Do(func() {
		var err error
		flutterBin, err = exec.LookPath("flutter")
		if err != nil {
			log.Errorf("Failed to lookup 'flutter' executable: %s. Please install flutter or add `--docker` to run the Hover command in Docker container.\nhttps://flutter.dev/docs/get-started/install", err)
			os.Exit(1)
		}
	})
	return flutterBin
}

var (
	gitBin     string
	gitBinOnce sync.Once
)

func GitBin() string {
	goBinOnce.Do(func() {
		var err error
		gitBin, err = exec.LookPath("git")
		if err != nil {
			log.Warnf("Failed to lookup 'git' executable: %s.", err)
		}
	})
	return gitBin
}
