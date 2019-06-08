package pubspec

import (
	"os"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Pubspec contains the parsed contents of pubspec.yaml
type Pubspec struct {
	Name         string
	Dependencies map[string]interface{}
}

// ReadLocal reads pubspec.yaml in the current working directory.
func ReadLocal() (*Pubspec, error) {
	p := &Pubspec{}
	file, err := os.Open("pubspec.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("no pubspec.yaml file found")

		}
		return nil, errors.Wrap(err, "failed to open pubspec.yaml")
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode pubspec.yaml")
	}
	return p, nil
}
