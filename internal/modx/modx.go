package modx

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
)

// Open the go.mod file in the given directory
func Open(dir string) (m *modfile.File, err error) {
	goModPath := filepath.Join(dir, "go.mod")

	goModBytes, err := ioutil.ReadFile(goModPath)
	if err != nil && !os.IsNotExist(err) {
		return m, errors.Wrapf(err, "failed to read the 'go.mod' file: %v", goModPath)
	}

	if m, err = modfile.Parse(goModPath, goModBytes, nil); err != nil {
		return m, errors.Wrapf(err, "failed to read the 'go.mod' file: %v", goModPath)
	}

	return m, nil
}
