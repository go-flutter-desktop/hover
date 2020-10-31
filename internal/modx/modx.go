package modx

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

// Open the go.mod file in the given directory
func Open(dir string) (m *modfile.File, err error) {
	if dir, err = FindModuleRoot(dir); err != nil {
		return m, err
	}

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

// Version locates the module version for the given import path.
// returns zero version if none are found.
// Version differs from find in that it returns the version in use
// vs the version required.
func Version(m *modfile.File, path string) module.Version {
	for _, pkg := range m.Replace {
		if pkg.Old.Path == path {
			return pkg.New
		}
	}

	for _, pkg := range m.Require {
		if pkg.Mod.Path == path {
			return pkg.Mod
		}
	}

	return module.Version{}
}

// Find locates the versions for the given import path.
// returns zero version if none are found.
func Find(m *modfile.File, path string) module.Version {
	for _, pkg := range m.Replace {
		if pkg.Old.Path == path {
			return pkg.Old
		}
	}

	for _, pkg := range m.Require {
		if pkg.Mod.Path == path {
			return pkg.Mod
		}
	}

	return module.Version{}
}

// RemoveModule drop a module from go.mod entirely.
func RemoveModule(m *modfile.File, path string) error {
	for v := Find(m, path); v.Path != ""; v = Find(m, path) {
		err := m.DropReplace(v.Path, v.Version)
		if err != nil {
			return err
		}

		err = m.DropRequire(v.Path)
		if err != nil {
			return err
		}
	}

	return nil
}

// Mutate the go.mod file in the given directory.
func Mutate(dir string, mutation func(*modfile.File) error) (err error) {
	var mod *modfile.File

	if mod, err = Open(dir); err != nil {
		return err
	}

	if err = mutation(mod); err != nil {
		return err
	}

	return Replace(dir, mod)
}

// Replace the go.mod file in the given directory.
func Replace(dir string, m *modfile.File) (err error) {
	m.Cleanup()

	if dir, err = FindModuleRoot(dir); err != nil {
		return err
	}

	goModPath := filepath.Join(dir, "go.mod")

	out, err := m.Format()
	if err != nil {
		return errors.Wrapf(err, "failed to format the 'go.mod' file: %s", goModPath)
	}

	info, err := os.Stat(goModPath)
	if err != nil {
		return errors.Wrapf(err, "failed to stat the 'go.mod' file: %s", goModPath)
	}

	err = ioutil.WriteFile(goModPath, out, info.Mode().Perm())
	if err != nil {
		return errors.Wrapf(err, "failed to update the 'go.mod' file: %s", goModPath)
	}

	return nil
}

// Print the modfile.
func Print(m *modfile.File) (s string, err error) {
	m.Cleanup()
	out, err := m.Format()
	if err != nil {
		return "", errors.Wrapf(err, "failed to format the 'go.mod' file")
	}

	return string(out), nil
}

// FindModuleRoot pulled from: https://github.com/golang/go/blob/88e564edb13f1596c12ad16d5fd3c7ac7deac855/src/cmd/dist/build.go#L1595
func FindModuleRoot(dir string) (cleaned string, err error) {
	if dir == "" {
		return "", errors.New("cannot located go.mod from a blank directory path")
	}

	if cleaned, err = filepath.Abs(filepath.Clean(dir)); err != nil {
		return "", errors.Wrap(err, "failed to determined absolute path to directory")
	}

	// Look for enclosing go.mod.
	for {
		gomod := filepath.Join(cleaned, "go.mod")
		if fi, err := os.Stat(gomod); err == nil && !fi.IsDir() {
			return cleaned, nil
		}

		d := filepath.Dir(cleaned)

		if d == cleaned {
			break
		}

		cleaned = d
	}

	return "", errors.Errorf("go.mod not found: %s", dir)
}
