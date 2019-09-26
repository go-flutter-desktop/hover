package pubspec

import (
	"github.com/go-flutter-desktop/hover/internal/log"
	"gopkg.in/yaml.v2"
	"os"
)

type PubSpec struct {
	Name         string
	Description  string
	Version      string
	Author       string
	Dependencies map[string]interface{}
}

var pubspec = PubSpec{}

func GetPubSpec() PubSpec {
	{
		if pubspec.Name == "" {
			file, err := os.Open("pubspec.yaml")
			if err != nil {
				if os.IsNotExist(err) {
					log.Errorf("Error: No pubspec.yaml file found.")
					goto Fail
				}
				log.Errorf("Failed to open pubspec.yaml: %v", err)
				os.Exit(1)
			}
			defer file.Close()

			err = yaml.NewDecoder(file).Decode(&pubspec)
			if err != nil {
				log.Errorf("Failed to decode pubspec.yaml: %v", err)
				goto Fail
			}
			if _, exists := pubspec.Dependencies["flutter"]; !exists {
				log.Errorf("Missing `flutter` in pubspec.yaml dependencies list.")
				goto Fail
			}
		}

		return pubspec
	}

Fail:
	log.Errorf("This command should be run from the root of your Flutter project.")
	os.Exit(1)
	return PubSpec{}
}
