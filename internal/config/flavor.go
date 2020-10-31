package config

import (
	"github.com/go-flutter-desktop/hover/internal/log"
	"os"
)

var hoverYaml string

func GetHoverFlavorYaml() string {
	return hoverYaml
}

func SetDefaultFlavorFile() {
	hoverYaml = "hover.yaml"
}

func SetHoverFlavor(flavor string) {
	hoverYaml = "hover-" + flavor + ".yaml"
	assertYamlFileExists(hoverYaml)
}

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
