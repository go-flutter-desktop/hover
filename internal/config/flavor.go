package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
)

var hoverYaml string

// GetHoverFlavorYaml returns the Hover yaml file
func GetHoverFlavorYaml() string {
	if len(hoverYaml) == 0 {
		hoverYaml = "hover.yaml"
	}
	return hoverYaml
}

// SetHoverFlavor sets the user defined hover flavor.
// eg. hover-develop.yaml, hover-staging.yaml, etc.
func SetHoverFlavor(flavor string) {
	hoverYaml = fmt.Sprintf("hover-%s.yaml", flavor)
	assertYamlFileExists(hoverYaml)
}

// assertYamlFileExists checks to see if the user defined yaml file exists
func assertYamlFileExists(yamlFile string) {
	_, err := os.Stat(filepath.Join(build.BuildPath, yamlFile))
	if os.IsNotExist(err) {
		log.Warnf("Hover Yaml file \"%s\" not found.", yamlFile)
		os.Exit(1)
	}
	if err != nil {
		log.Errorf("Failed to stat %s: %v\n", yamlFile, err)
		os.Exit(1)
	}
}
