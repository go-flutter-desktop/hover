package pubspec

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/go-flutter-desktop/hover/internal/log"
)

// PubSpec contains the parsed contents of pubspec.yaml
type PubSpec struct {
	Name         string
	Description  string
	Version      string
	Author       string
	Dependencies map[string]interface{}
	Flutter      map[string]interface{}
}

var pubspec = PubSpec{}

// GetPubSpec returns the working directory pubspec.yaml as a PubSpec
func GetPubSpec() PubSpec {
	if pubspec.Name == "" {
		pub, err := ReadPubSpecFile("pubspec.yaml")
		if err != nil {
			log.Errorf("%v", err)
			log.Errorf("This command should be run from the root of your Flutter project.")
			os.Exit(1)
		}
		pubspec = *pub
	}
	return pubspec
}

// ReadPubSpecFile reads a .yaml file at a path and return a correspond
// PubSpec struct
func ReadPubSpecFile(pubSpecPath string) (*PubSpec, error) {
	file, err := os.Open(pubSpecPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(err, "Error: No pubspec.yaml file found")
		}
		return nil, errors.Wrap(err, "Failed to open pubspec.yaml")
	}
	defer file.Close()

	var pub PubSpec
	err = yaml.NewDecoder(file).Decode(&pub)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode pubspec.yaml")
	}
	// avoid checking for the flutter dependencies for out of ws directories
	if pubSpecPath != "pubspec.yaml" {
		return &pub, nil
	}
	if _, exists := pub.Dependencies["flutter"]; !exists {
		return nil, errors.New(fmt.Sprintf("Missing `flutter` in %s dependencies list", pubSpecPath))
	}
	return &pub, nil
}
