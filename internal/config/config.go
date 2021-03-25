package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/go-flutter-desktop/hover/internal/androidmanifest"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
)

// BuildTargetDefault Default build target file
const BuildTargetDefault = "lib/main_desktop.dart"

// BuildEngineDefault Default go-flutter engine version
const BuildEngineDefault = ""

// BuildOpenGlVersionDefault Default OpenGL version for go-flutter
const BuildOpenGlVersionDefault = "3.3"

// Config contains the parsed contents of hover.yaml
type Config struct {
	ApplicationName  string `yaml:"application-name"`
	ExecutableName   string `yaml:"executable-name"`
	PackageName      string `yaml:"package-name"`
	OrganizationName string `yaml:"organization-name"`
	License          string
	Target           string
	BranchREMOVED    string `yaml:"branch"`
	CachePathREMOVED string `yaml:"cache-path"`
	OpenGL           string
	Engine           string `yaml:"engine-version"`
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

func (c Config) GetOrganizationName() string {
	if len(c.OrganizationName) == 0 {
		PrintMissingField("organization-name", "go/hover.yaml", c.OrganizationName)
		// It would be nicer to not load a value from the AndroidManifest.xml and instead define a default value here,
		// but then older apps might break so for compatibility reasons it's done this way.
		c.OrganizationName = androidmanifest.AndroidOrganizationName()
	}
	return c.OrganizationName
}

func (c Config) GetLicense() string {
	if len(c.License) == 0 {
		c.License = "NOASSERTION"
		PrintMissingField("license", "go/hover.yaml", c.License)
	}
	return c.License
}

var (
	config         Config
	configLoadOnce sync.Once
)

// GetConfig returns the working directory hover.yaml as a Config
func GetConfig() Config {
	configLoadOnce.Do(func() {
		var err error
		hoverYaml := GetHoverFlavorYaml()
		config, err = ReadConfigFile(filepath.Join(build.BuildPath, hoverYaml))
		if err != nil {
			if os.IsNotExist(errors.Cause(err)) {
				// TODO: Add a solution for the user. Perhaps we can let `hover
				// init` write missing files when ran on an existing project.
				// https://github.com/go-flutter-desktop/hover/pull/121#pullrequestreview-408680348
				log.Warnf("Missing config: %v", err)
				return
			}
			log.Errorf("Failed to load config: %v", err)
			os.Exit(1)
		}

		if config.CachePathREMOVED != "" {
			log.Errorf("The hover.yaml field 'cache-path' is not used anymore. Remove it from your hover.yaml and use --cache-path instead.")
			os.Exit(1)
		}
		if config.BranchREMOVED != "" {
			log.Errorf("The hover.yaml field 'branch' is not used anymore. Remove it from your hover.yaml and use --branch instead.")
			os.Exit(1)
		}
	})
	return config
}

// ReadConfigFile reads a .yaml file at a path and return a correspond Config
// struct
func ReadConfigFile(configPath string) (Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, errors.Wrap(err, "file hover.yaml not found")
		}
		return Config{}, errors.Wrap(err, "failed to open hover.yaml")
	}
	defer file.Close()

	var config Config
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return Config{}, errors.Wrap(err, "failed to decode hover.yaml")
	}
	return config, nil
}

func PrintMissingField(name, file, def string) {
	log.Warnf("Missing/Empty `%s` field in %s. Please add it or otherwise you may publish your app with a wrong %s. Continuing with `%s` as a placeholder %s.", name, file, name, def, name)
}
