package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

const standaloneImplementationListAPI = "https://raw.githubusercontent.com/go-flutter-desktop/plugins/master/list.json"

var (
	listAllPluginDependencies bool
	tidyPurge                 bool
	dryRun                    bool
	reImport                  bool
)

func init() {
	pluginTidyCmd.Flags().BoolVar(&tidyPurge, "purge", false, "Remove all go platform plugins imports from the project.")
	pluginListCmd.Flags().BoolVarP(&listAllPluginDependencies, "all", "a", false, "List all platform plugins dependencies, even the one have no go-flutter support")
	pluginGetCmd.Flags().BoolVar(&reImport, "force", false, "Re-import already imported plugins.")

	pluginCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Perform a trial run with no changes made.")

	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginGetCmd)
	pluginCmd.AddCommand(pluginTidyCmd)
	rootCmd.AddCommand(pluginCmd)
}

var pluginCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Tools for plugins",
	Long:  "A collection of commands to help with finding/importing go-flutter implementations of plugins.",
}

// PubSpecLock contains the parsed contents of pubspec.lock
type PubSpecLock struct {
	Packages map[string]PubDep
}

// PubDep contains one entry of the pubspec.lock yaml list
type PubDep struct {
	Dependency  string
	Description interface{}
	Source      string
	Version     string

	// Fields set by hover
	name    string
	android bool
	ios     bool
	desktop bool
	// optional description values
	path string // correspond to the path field in lock file
	host string // correspond to the host field in lock file
	// contain a import.go.tmpl file used for import
	autoImport bool
	// the path/URL to the go code of the plugin is stored
	pluginGoSource string
	// whether or not the go plugin source code is located on another VCS repo.
	standaloneImpl bool
}

func (p PubDep) imported() bool {
	pluginImportOutPath := filepath.Join(build.BuildPath, "cmd", fmt.Sprintf("import-%s-plugin.go", p.name))
	if _, err := os.Stat(pluginImportOutPath); err == nil {
		return true
	}
	return false
}

func (p PubDep) platforms() []string {
	var platforms []string
	if p.android {
		platforms = append(platforms, "android")
	}
	if p.ios {
		platforms = append(platforms, "ios")
	}
	if p.desktop {
		platforms = append(platforms, build.BuildPath)
	}
	return platforms
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List golang platform plugins in the application",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("does not take arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterProject()
		dependencyList, err := listPlatformPlugin()
		if err != nil {
			log.Errorf("%v", err)
			os.Exit(1)
		}

		var hasNewPlugin bool
		var hasPlugins bool
		for _, dep := range dependencyList {
			if !(dep.desktop || listAllPluginDependencies) {
				continue
			}

			if hasPlugins {
				fmt.Println("")
			}
			hasPlugins = true

			fmt.Printf("     - %s\n", dep.name)
			fmt.Printf("         version:   %s\n", dep.Version)
			fmt.Printf("         platforms: [%s]\n", strings.Join(dep.platforms(), ", "))
			if dep.desktop {
				if dep.standaloneImpl {
					fmt.Printf("         source:    This go plugin isn't maintained by the official plugin creator.\n")
				}
				if dep.imported() {
					fmt.Println("         import:    [OK] The plugin is already imported in the project.")
					continue
				}
				if dep.autoImport || dep.standaloneImpl {
					hasNewPlugin = true
					fmt.Println("         import:    [Missing] The plugin can be imported by hover.")
				} else {
					fmt.Println("         import:    [Manual import] The plugin is missing the import.go.tmpl file required for hover import.")
				}
				if dep.path != "" {
					fmt.Printf("         dev:       Plugin replaced in go.mod to path: '%s'\n", dep.path)
				}
			}
		}
		if hasNewPlugin {
			log.Infof(fmt.Sprintf("run `%s` to import the missing plugins!", log.Au().Magenta("hover plugins get")))
		}
	},
}

var pluginTidyCmd = &cobra.Command{
	Use:   "tidy",
	Short: "Removes unused platform plugins.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("does not take arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterProject()
		assertHoverInitialized()

		desktopCmdPath := filepath.Join(build.BuildPath, "cmd")
		dependencyList, err := listPlatformPlugin()
		if err != nil {
			log.Errorf("%v", err)
			os.Exit(1)
		}

		importedPlugins, err := ioutil.ReadDir(desktopCmdPath)
		if err != nil {
			log.Errorf("Failed to search for plugins: %v", err)
			os.Exit(1)
		}

		for _, f := range importedPlugins {
			isPlugin := strings.HasPrefix(f.Name(), "import-")
			isPlugin = isPlugin && strings.HasSuffix(f.Name(), "-plugin.go")

			if isPlugin {
				pluginName := strings.TrimPrefix(f.Name(), "import-")
				pluginName = strings.TrimSuffix(pluginName, "-plugin.go")

				pluginInUse := false
				for _, dep := range dependencyList {
					if dep.name == pluginName {
						// plugin in pubspec.lock
						pluginInUse = true
						break
					}
				}

				if !pluginInUse || tidyPurge {
					if dryRun {
						fmt.Printf("       plugin: [%s] can be removed\n", pluginName)
						continue
					}
					pluginImportPath := filepath.Join(desktopCmdPath, f.Name())

					// clean-up go.mod
					pluginImportStr, _ := readPluginGoImport(pluginImportPath, pluginName)
					// Delete the 'replace' and 'require' import strings from go.mod.
					// Not mission critical, if the plugins not correctly removed from
					// the go.mod file, the project still works and the plugin is
					// successfully removed from the flutter.Application.
					if err != nil || pluginImportStr == "" {
						log.Warnf("Couldn't clean the '%s' plugin from the 'go.mod' file. Error: %v", pluginName, err)
					} else {
						fileutils.RemoveLinesFromFile(filepath.Join(build.BuildPath, "go.mod"), pluginImportStr)
					}

					// remove import file
					err = os.Remove(pluginImportPath)
					if err != nil {
						log.Warnf("Couldn't remove plugin %s: %v", pluginName, err)
						continue
					}
					fmt.Printf("       plugin: [%s] removed\n", pluginName)
				}
			}
		}
	},
}

var pluginGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Imports missing platform plugins in the application",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("does not take arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterProject()
		assertHoverInitialized()
		hoverPluginGet(false)
	},
}

func hoverPluginGet(dryRun bool) bool {
	dependencyList, err := listPlatformPlugin()
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

	for _, dep := range dependencyList {

		if !dep.desktop {
			continue
		}

		if !dep.autoImport {
			fmt.Printf("       plugin: [%s] couldn't be imported, check the plugin's README for manual instructions\n", dep.name)
			continue
		}

		if dryRun {
			if dep.imported() {
				fmt.Printf("       plugin: [%s] can be updated\n", dep.name)
			} else {
				fmt.Printf("       plugin: [%s] can be imported\n", dep.name)
			}
			continue
		}

		pluginImportOutPath := filepath.Join(build.BuildPath, "cmd", fmt.Sprintf("import-%s-plugin.go", dep.name))

		if dep.imported() && !reImport {
			pluginImportStr, err := readPluginGoImport(pluginImportOutPath, dep.name)
			if err != nil {
				log.Warnf("Couldn't read the plugin '%s' import URL", dep.name)
				log.Warnf("Fallback to the latest version installed.")
				continue
			}

			if !goGetModuleSuccess(pluginImportStr, dep.Version) {
				log.Warnf("Couldn't download version '%s' of plugin '%s'", dep.Version, dep.name)
				log.Warnf("Fallback to the latest version installed.")
				continue
			}

			fmt.Printf("       plugin: [%s] updated\n", dep.name)
			continue
		}

		if dep.standaloneImpl {
			fileutils.DownloadFile(dep.pluginGoSource, pluginImportOutPath)
		} else {
			autoImportTemplatePath := filepath.Join(dep.pluginGoSource, "import.go.tmpl")
			fileutils.CopyFile(autoImportTemplatePath, pluginImportOutPath)

			pluginImportStr, err := readPluginGoImport(pluginImportOutPath, dep.name)
			if err != nil {
				log.Warnf("Couldn't read the plugin '%s' import URL", dep.name)
				log.Warnf("Fallback to the latest version available on github.")
				continue
			}

			// if remote plugin, get the correct version
			if dep.path == "" {
				if !goGetModuleSuccess(pluginImportStr, dep.Version) {
					log.Warnf("Couldn't download version '%s' of plugin '%s'", dep.Version, dep.name)
					log.Warnf("Fallback to the latest version available on github.")
				}
			}

			// if local plugin
			if dep.path != "" {
				path, err := filepath.Abs(filepath.Join(dep.path, "go"))
				if err != nil {
					log.Errorf("Failed to resolve absolute path for plugin '%s': %v", dep.name, err)
					os.Exit(1)
				}
				fileutils.AddLineToFile(filepath.Join(build.BuildPath, "go.mod"), fmt.Sprintf("replace %s => %s", pluginImportStr, path))
			}

			fmt.Printf("       plugin: [%s] imported\n", dep.name)
		}
	}

	return len(dependencyList) != 0
}

func listPlatformPlugin() ([]PubDep, error) {
	onlineList, err := fetchStandaloneImplementationList()
	if err != nil {
		log.Warnf("Warning, couldn't read the online plugin list: %v", err)
	}

	pubcachePath, err := findPubcachePath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to find path for pub-cache")
	}

	var list []PubDep
	pubLock, err := readPubSpecLock()

	for name, entry := range pubLock.Packages {
		entry.name = name

		switch i := entry.Description.(type) {
		case string:
			if i == "flutter" {
				continue
			}
		case map[interface{}]interface{}:
			if value, ok := i["path"]; ok {
				entry.path = value.(string)
			}
			if value, ok := i["url"]; ok {
				url, err := url.Parse(value.(string))
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse URL from string %s"+value.(string))
				}
				entry.host = url.Host
			}
		}

		pluginPath := filepath.Join(pubcachePath, "hosted", entry.host, entry.name+"-"+entry.Version)
		if entry.path != "" {
			pluginPath = entry.path
		}

		pluginPubspecPath := filepath.Join(pluginPath, "pubspec.yaml")
		pluginPubspec, err := pubspec.ReadPubSpecFile(pluginPubspecPath)
		if err != nil {
			continue
		}

		// Non plugin package are likely to contain android/ios folders (even
		// through they aren't used).
		// To check if the package is really a platform plugin, we need to read
		// the pubspec.yaml file. If he contains a Flutter/plugin entry, then
		// it's a platform plugin.
		if _, ok := pluginPubspec.Flutter["plugin"]; !ok {
			continue
		}

		detectPlatformPlugin := func(platform string) (bool, error) {
			platformPath := filepath.Join(pluginPath, platform)
			stat, err := os.Stat(platformPath)
			if err != nil {
				if os.IsNotExist(err) {
					return false, nil
				}
				return false, errors.Wrapf(err, "failed to stat %s", platformPath)
			}
			return stat.IsDir(), nil
		}

		entry.android, err = detectPlatformPlugin("android")
		if err != nil {
			return nil, err
		}
		entry.ios, err = detectPlatformPlugin("ios")
		if err != nil {
			return nil, err
		}
		entry.desktop, err = detectPlatformPlugin(build.BuildPath)
		if err != nil {
			return nil, err
		}

		if entry.desktop {
			entry.pluginGoSource = filepath.Join(pluginPath, build.BuildPath)
			autoImportTemplate := filepath.Join(entry.pluginGoSource, "import.go.tmpl")
			_, err := os.Stat(autoImportTemplate)
			entry.autoImport = true
			if err != nil {
				entry.autoImport = false
				if !os.IsNotExist(err) {
					return nil, errors.Wrapf(err, "failed to stat %s", autoImportTemplate)
				}
			}
		} else {
			// check if the plugin is available in github.com/go-flutter-desktop/plugins
			for _, plugin := range onlineList {
				if entry.name == plugin.Name {
					entry.desktop = true
					entry.standaloneImpl = true
					entry.autoImport = true
					entry.pluginGoSource = plugin.ImportFile
					break
				}
			}
		}

		list = append(list, entry)

	}
	return list, nil
}

// readLocal reads pubspec.lock in the current working directory.
func readPubSpecLock() (*PubSpecLock, error) {
	p := &PubSpecLock{}
	file, err := os.Open("pubspec.lock")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("no pubspec.lock file found")

		}
		return nil, errors.Wrap(err, "failed to open pubspec.lock")
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode pubspec.lock")
	}
	return p, nil
}

func readPluginGoImport(pluginImportOutPath, pluginName string) (string, error) {
	pluginImportBytes, err := ioutil.ReadFile(pluginImportOutPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	re := regexp.MustCompile(fmt.Sprintf(`\s+%s\s"(\S*)"`, pluginName))

	match := re.FindStringSubmatch(string(pluginImportBytes))
	if len(match) < 2 {
		err = errors.New("Failed to parse the import path, plugin name in the import must have been changed")
		return "", err
	}
	return match[1], nil
}

type onlineList struct {
	List []StandaloneImplementation `json:"standaloneImplementation"`
}

// StandaloneImplementation contains the go-flutter compatible plugins that
// aren't merged into original VSC repo.
type StandaloneImplementation struct {
	Name       string `json:"name"`
	ImportFile string `json:"importFile"`
}

func fetchStandaloneImplementationList() ([]StandaloneImplementation, error) {
	remoteList := &onlineList{}

	client := http.Client{
		Timeout: time.Second * 20, // Maximum of 10 secs
	}

	req, err := http.NewRequest(http.MethodGet, standaloneImplementationListAPI, nil)
	if err != nil {
		return remoteList.List, err
	}

	res, err := client.Do(req)
	if err != nil {
		return remoteList.List, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return remoteList.List, err
	}

	if res.StatusCode != 200 {
		return remoteList.List, errors.New(strings.TrimRight(string(body), "\r\n"))
	}

	err = json.Unmarshal(body, remoteList)
	if err != nil {
		return remoteList.List, err
	}
	return remoteList.List, nil
}

// goGetModuleSuccess updates a module at a version, if it fails, return false.
func goGetModuleSuccess(pluginImportStr, version string) bool {
	cmdGoGetU := exec.Command(goBin, "get", "-u", pluginImportStr+"@v"+version)
	cmdGoGetU.Dir = filepath.Join(build.BuildPath)
	cmdGoGetU.Env = append(os.Environ(),
		"GOPROXY=direct", // github.com/golang/go/issues/32955 (allows '/' in branch name)
		"GO111MODULE=on",
	)
	cmdGoGetU.Stderr = os.Stderr
	cmdGoGetU.Stdout = os.Stdout
	return cmdGoGetU.Run() == nil
}
