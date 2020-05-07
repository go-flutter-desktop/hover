package packaging

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"

	"github.com/otiai10/copy"

	"github.com/go-flutter-desktop/hover/internal/androidmanifest"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/config"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

var packagingPath = filepath.Join(build.BuildPath, "packaging")

func packagingFormatPath(packagingFormat string) string {
	directoryPath, err := filepath.Abs(filepath.Join(packagingPath, packagingFormat))
	if err != nil {
		log.Errorf("Failed to resolve absolute path for %s directory: %v", packagingFormat, err)
		os.Exit(1)
	}
	return directoryPath
}

func createPackagingFormatDirectory(packagingFormat string) {
	if _, err := os.Stat(packagingFormatPath(packagingFormat)); !os.IsNotExist(err) {
		log.Errorf("A file or directory named `%s` already exists. Cannot continue packaging init for %s.", packagingFormat, packagingFormat)
		os.Exit(1)
	}
	err := os.MkdirAll(packagingFormatPath(packagingFormat), 0775)
	if err != nil {
		log.Errorf("Failed to create %s directory %s: %v", packagingFormat, packagingFormatPath(packagingFormat), err)
		os.Exit(1)
	}
}

func getTemporaryBuildDirectory(projectName string, packagingFormat string) string {
	tmpPath, err := ioutil.TempDir("", "hover-build-"+projectName+"-"+packagingFormat)
	if err != nil {
		log.Errorf("Couldn't get temporary build directory: %v", err)
		os.Exit(1)
	}
	return tmpPath
}

type packagingTask struct {
	packagingFormatName            string                                                                                                                        // Name of the packaging format: OS-TYPE
	dependsOn                      map[*packagingTask]string                                                                                                     // Packaging tasks this task depends on
	templateFiles                  map[string]string                                                                                                             // Template files to copy over on init
	executableFiles                []string                                                                                                                      // Files that should be executable
	linuxDesktopFileExecutablePath string                                                                                                                        // Path of the executable for linux .desktop file (only set on linux)
	linuxDesktopFileIconPath       string                                                                                                                        // Path of the icon for linux .desktop file (only set on linux)
	generateBuildFiles             func(packageName, path string)                                                                                                // Generate dynamic build files. Operates in the temporary directory
	generateInitFiles              func(packageName, path string)                                                                                                // Generate dynamic init files
	extraTemplateData             func(packageName, path string) map[string]string                                                                              // Update the template data on build. This is used for inserting values that are generated on init
	flutterBuildOutputDirectory    string                                                                                                                        // Path to copy the build output of the app to. Operates in the temporary directory
	packagingFunction              func(tmpPath, applicationName, strippedApplicationName, packageName, executableName, version, release string) (string, error) // Function that actually packages the app. Needs to check for OS specific tools etc. . Returns the path of the packaged file
	skipAssertInitialized          bool                                                                                                                          // Set to true when a task doesn't need to be initialized.
	requiredTools                  map[string][]string                                                                                                           // Map of list of tools required to package per OS
}

func (t *packagingTask) AssertSupported() {
	for task := range t.dependsOn {
		task.AssertSupported()
	}
	if _, osIsSupported := t.requiredTools[runtime.GOOS]; !osIsSupported {
		log.Errorf("Packaging %s is not supported on %s", t.packagingFormatName, runtime.GOOS)
		log.Errorf("To still package %s on %s you need to run hover with the `--docker` flag.", t.packagingFormatName, runtime.GOOS)
		os.Exit(1)
	}
	var unavailableTools []string
	for _, tool := range t.requiredTools[runtime.GOOS] {
		_, err := exec.LookPath(tool)
		if err != nil {
			unavailableTools = append(unavailableTools, tool)
		}
	}
	if len(unavailableTools) > 0 {
		log.Errorf("To package %s these tools are required: %s", t.packagingFormatName, strings.Join(unavailableTools, ","))
		log.Errorf("To still package %s without the required tools installed you need to run hover with the `--docker` flag.", t.packagingFormatName)
		os.Exit(1)
	}
}

var templateData map[string]string
var once sync.Once

var (
	projectName             string
	version                 string
	release                 string
	description             string
	organizationName        string
	author                  string
	applicationName         string
	strippedApplicationName string
	executableName          string
	packageName             string
	license                 string
)

func (t *packagingTask) initData(fullVersion string) map[string]string {
	once.Do(func() {
		projectName = pubspec.GetPubSpec().Name
		version = strings.Split(fullVersion, "+")[0]
		if strings.Contains(fullVersion, "+") {
			release = strings.Split(fullVersion, "+")[1]
		} else {
			release = strings.ReplaceAll(fullVersion, ".", "")
		}
		description = pubspec.GetPubSpec().Description
		organizationName = androidmanifest.AndroidOrganizationName()
		author = pubspec.GetPubSpec().GetAuthor()
		applicationName = config.GetConfig().GetApplicationName(projectName)
		strippedApplicationName = strings.ReplaceAll(applicationName, " ", "_")
		executableName = config.GetConfig().GetExecutableName(projectName)
		packageName = config.GetConfig().GetPackageName(projectName)
		license = config.GetConfig().GetLicense()
		templateData = map[string]string{
			"projectName":             projectName,
			"version":                 version,
			"release":                 release,
			"description":             description,
			"organizationName":        organizationName,
			"author":                  author,
			"applicationName":         applicationName,
			"strippedApplicationName": strippedApplicationName,
			"executableName":          executableName,
			"packageName":             packageName,
			"license":                 license,
		}
		templateData["iconPath"] = executeStringTemplate(t.linuxDesktopFileIconPath, templateData)
		templateData["executablePath"] = executeStringTemplate(t.linuxDesktopFileExecutablePath, templateData)
	})
	return templateData
}

func (t *packagingTask) Name() string {
	return strings.SplitN(t.packagingFormatName, "-", 2)[1]
}

func (t *packagingTask) Init() {
	t.init(false)
}

func (t *packagingTask) init(ignoreAlreadyExists bool) {
	for task := range t.dependsOn {
		task.init(true)
	}
	if !t.IsInitialized() {
		createPackagingFormatDirectory(t.packagingFormatName)
		dir := packagingFormatPath(t.packagingFormatName)
		for sourceFile, destinationFile := range t.templateFiles {
			destinationFile = filepath.Join(dir, destinationFile)
			err := os.MkdirAll(filepath.Dir(destinationFile), 0775)
			if err != nil {
				log.Errorf("Failed to create directory %s: %v", filepath.Dir(destinationFile), err)
				os.Exit(1)
			}
			fileutils.CopyAsset(fmt.Sprintf("packaging/%s", sourceFile), destinationFile, fileutils.AssetsBox())
		}
		if t.generateInitFiles != nil {
			log.Infof("Generating dynamic init files")
			t.generateInitFiles(config.GetConfig().GetPackageName(projectName), dir)
		}
		log.Infof("go/packaging/%s has been created. You can modify the configuration files and add it to git.", t.packagingFormatName)
		log.Infof(fmt.Sprintf("You now can package the %s using `%s`", strings.Split(t.packagingFormatName, "-")[0], log.Au().Magenta("hover build "+t.packagingFormatName)))
	} else if !ignoreAlreadyExists {
		log.Errorf("%s is already initialized for packaging.", t.packagingFormatName)
		os.Exit(1)
	}
}

func (t *packagingTask) Pack(fullVersion string) {
	t.initData(fullVersion)
	t.pack(fullVersion)
}

func (t *packagingTask) pack(fullVersion string) {
	for key, value := range t.extraTemplateData(packageName, packagingFormatPath(t.packagingFormatName)) {
		templateData[key] = value
	}
	for task := range t.dependsOn {
		task.pack(fullVersion)
	}
	tmpPath := getTemporaryBuildDirectory(projectName, t.packagingFormatName)
	defer func() {
		err := os.RemoveAll(tmpPath)
		if err != nil {
			log.Errorf("Could not remove temporary build directory: %v", err)
			os.Exit(1)
		}
	}()
	log.Infof("Packaging %s in %s", strings.Split(t.packagingFormatName, "-")[1], tmpPath)

	if t.flutterBuildOutputDirectory != "" {
		err := copy.Copy(build.OutputDirectoryPath(strings.Split(t.packagingFormatName, "-")[0]), executeStringTemplate(filepath.Join(tmpPath, t.flutterBuildOutputDirectory), templateData))
		if err != nil {
			log.Errorf("Could not copy build folder: %v", err)
			os.Exit(1)
		}
	}
	for task, destination := range t.dependsOn {
		err := copy.Copy(build.OutputDirectoryPath(task.packagingFormatName), filepath.Join(tmpPath, destination))
		if err != nil {
			log.Errorf("Could not copy build folder of %s: %v", task.packagingFormatName, err)
			os.Exit(1)
		}
	}
	fileutils.CopyTemplateDir(packagingFormatPath(t.packagingFormatName), filepath.Join(tmpPath), templateData)
	if t.generateBuildFiles != nil {
		log.Infof("Generating dynamic build files")
		t.generateBuildFiles(packageName, tmpPath)
	}

	for _, file := range t.executableFiles {
		err := os.Chmod(executeStringTemplate(filepath.Join(tmpPath, file), templateData), 0777)
		if err != nil {
			log.Errorf("Failed to change file permissions for %s file: %v", file, err)
			os.Exit(1)
		}
	}

	err := os.RemoveAll(build.OutputDirectoryPath(t.packagingFormatName))
	log.Printf("Cleaning the build directory")
	if err != nil {
		log.Errorf("Failed to clean output directory %s: %v", build.OutputDirectoryPath(t.packagingFormatName), err)
		os.Exit(1)
	}

	relativeOutputFilePath, err := t.packagingFunction(tmpPath, applicationName, strippedApplicationName, packageName, executableName, version, release)
	if err != nil {
		log.Errorf("%v", err)
		log.Warnf("Packaging is very experimental and has mostly been tested on Linux.")
		log.Infof("Please open an issue at https://github.com/go-flutter-desktop/go-flutter/issues/new?template=BUG.md")
		log.Infof("with the log and a reproducible example if possible. You may also zip your app code")
		log.Infof("if you are comfortable with it (closed source etc.) and attach it to the issue.")
		os.Exit(1)
	}
	outputFileName := filepath.Base(relativeOutputFilePath)
	outputFilePath := filepath.Join(build.OutputDirectoryPath(t.packagingFormatName), outputFileName)
	err = copy.Copy(filepath.Join(tmpPath, relativeOutputFilePath), outputFilePath)
	if err != nil {
		log.Errorf("Could not move %s file: %v", outputFileName, err)
		os.Exit(1)
	}
}

func (t *packagingTask) AssertInitialized() {
	if t.skipAssertInitialized {
		return
	}
	if !t.IsInitialized() {
		log.Errorf("%s is not initialized for packaging. Please run `hover init-packaging %s` first.", t.packagingFormatName, t.packagingFormatName)
		os.Exit(1)
	}
}

func (t *packagingTask) IsInitialized() bool {
	_, err := os.Stat(packagingFormatPath(t.packagingFormatName))
	return !os.IsNotExist(err)
}

func executeStringTemplate(t string, data map[string]string) string {
	tmplFile, err := template.New("").Option("missingkey=error").Parse(t)
	if err != nil {
		log.Errorf("Failed to parse template string: %v\n", err)
		os.Exit(1)
	}
	var tmplBytes bytes.Buffer
	err = tmplFile.Execute(&tmplBytes, data)
	if err != nil {
		panic(err)
	}
	return tmplBytes.String()
}
