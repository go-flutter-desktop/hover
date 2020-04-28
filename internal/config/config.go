package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
)

// BuildTargetDefault Default build target file
const BuildTargetDefault = "lib/main_desktop.dart"

// BuildBranchDefault Default go-flutter branch
const BuildBranchDefault = ""

// BuildEngineDefault Default go-flutter engine version
const BuildEngineDefault = ""

// BuildOpenGlVersionDefault Default OpenGL version for go-flutter
const BuildOpenGlVersionDefault = "3.3"

// Config contains the parsed contents of hover.yaml
type Config struct {
	loaded          bool
	ApplicationName string `yaml:"application-name"`
	ExecutableName  string `yaml:"executable-name"`
	PackageName     string `yaml:"package-name"`
	License         string
	Target          string
	Branch          string
	CachePath       string `yaml:"cache-path"`
	OpenGL          string
	Engine          string `yaml:"engine-version"`
}

func (c Config) GetApplicationName(projectName string) string {
	if c.ApplicationName == "" {
		return projectName
	}
	return c.ApplicationName
}

func (c Config) GetExecutableName(projectName string) string {
	if c.ExecutableName == "" {
		return strings.ReplaceAll(projectName, " ", "")
	}
	return c.ExecutableName
}

func (c Config) GetPackageName(projectName string) string {
	if c.PackageName == "" {
		return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(projectName, "-", ""), "_", ""), " ", "")
	}
	return c.PackageName
}

func (c Config) GetLicense() string {
	if len(c.License) == 0 {
		c.License = "NOASSERTION"
		PrintMissingField("license", "go/hover.yaml", c.License)
	}
	return c.License
}

var config = Config{}

// GetConfig returns the working directory hover.yaml as a Config
func GetConfig() Config {
	if !config.loaded {
		c, err := ReadConfigFile(filepath.Join(build.BuildPath, "hover.yaml"))
		if err != nil {
			return config
		}
		config = *c
		config.loaded = true
	}
	return config
}

// ReadConfigFile reads a .yaml file at a path and return a correspond
// Config struct
func ReadConfigFile(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(err, "Warning: No hover.yaml file found")
		}
		return nil, errors.Wrap(err, "Failed to open hover.yaml")
	}
	defer file.Close()

	var config Config
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode hover.yaml")
	}
	return &config, nil
}

func PrintMissingField(name, file, def string) {
	log.Warnf("Missing/Empty `%s` field in %s. Please add it or otherwise you may publish your app with a wrong %s. Continuing with `%s` as a placeholder %s.", name, file, name, def, name)
}
