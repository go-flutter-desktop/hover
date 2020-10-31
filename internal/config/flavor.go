package config

import (
	"github.com/go-flutter-desktop/hover/internal/log"
	"os"
)

var hoverYaml string

// GetHoverFlavorYaml returns the Hover yaml file
func GetHoverFlavorYaml() string {
	return hoverYaml
}

// SetDefaultFlavorFile sets the default hover.yaml
func SetDefaultFlavorFile() {
	hoverYaml = "hover.yaml"
}

// SetHoverFlavor sets the user defined hover flavor.
// eg. hover-develop.yaml, hover-staging.yaml, etc.
func SetHoverFlavor(flavor string) {
	hoverYaml = "hover-" + flavor + ".yaml"
	assertYamlFileExists(hoverYaml)
}

// assertYamlFileExists checks to see if the user defined yaml file exists
func assertYamlFileExists(yamlFile string) {
	_, err := os.Stat(yamlFile)
	if os.IsNotExist(err) {
		log.Warnf("Hover Yaml file \"%s\" not found.", yamlFile)
		os.Exit(1)
	}
	if err != nil {
		log.Errorf("Failed to stat hover.yaml: %v\n", err)
		os.Exit(1)
	}
}
