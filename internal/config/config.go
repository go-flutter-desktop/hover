package config

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
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
	applicationName string `yaml:"application-name"`
	executableName  string `yaml:"executable-name"`
	packageName     string `yaml:"package-name"`
	license         string
	Target          string
	Branch          string
	CachePath       string `yaml:"cache-path"`
	OpenGL          string
	Engine          string `yaml:"engine-version"`
}

func (c Config) ApplicationName(projectName string) string {
	if c.applicationName == "" {
		return projectName
	}
	return c.applicationName
}

func (c Config) ExecutableName(projectName string) string {
	if c.executableName == "" {
		return strings.ReplaceAll(projectName, " ", "")
	}
	return c.executableName
}

func (c Config) PackageName(projectName string) string {
	if c.packageName == "" {
		return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(projectName, "-", ""), "_", ""), " ", "")
	}
	return c.packageName
}

func (c Config) License() string {
	if c.license == "" {
		log.Warnf("Missing/Empty `license` field in go/hover.yaml.")
		log.Warnf("Please add it otherwise you may publish your app with a wrong license.")
		log.Warnf("Continuing with `FIXME` as a placeholder license.")
		return "FIXME"
	}
	return c.license
}

func (c Config) Author() string {
	author := pubspec.GetPubSpec().Author
	if author == "" {
		log.Warnf("Missing author field in pubspec.yaml")
		log.Warnf("Please add the `author` field to your pubspec.yaml")
		u, err := user.Current()
		if err != nil {
			log.Errorf("Couldn't get current user: %v", err)
			os.Exit(1)
		}
		author = u.Username
		log.Printf("Using this username from system instead: %s", author)
	}
	return author
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
