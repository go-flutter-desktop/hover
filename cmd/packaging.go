package cmd

import (
	"fmt"
	"github.com/go-flutter-desktop/hover/internal/log"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

var packagingPath = filepath.Join(buildPath, "packaging")

func init() {
	initPackagingCmd.AddCommand(initLinuxSnapCmd)
	initPackagingCmd.AddCommand(initLinuxDebCmd)
	rootCmd.AddCommand(initPackagingCmd)
}

var initPackagingCmd = &cobra.Command{
	Use:   "init-packaging",
	Short: "Create configuration files for a packaging format",
}

var initLinuxSnapCmd = &cobra.Command{
	Use:   "linux-snap",
	Short: "Create configuration files for snap packaging",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := getPubSpec().Name
		assertHoverInitialized()

		initLinuxSnap(projectName)
	},
}

var initLinuxDebCmd = &cobra.Command{
	Use:   "linux-deb",
	Short: "Create configuration files for deb packaging",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := getPubSpec().Name
		assertHoverInitialized()

		initLinuxDeb(projectName)
	},
}

var linuxPackagingDependencies = []string{"libx11-6", "libxrandr2", "libxcursor1", "libxinerama1"}

func packagingFormatPath(packagingFormat string) string {
	directoryPath, err := filepath.Abs(filepath.Join(packagingPath, packagingFormat))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for %s directory: %v", packagingFormat, err)
		os.Exit(1)
	}
	return directoryPath
}

func createPackagingFormatDirectory(packagingFormat string) {
	if _, err := os.Stat(packagingFormatPath(packagingFormat)); !os.IsNotExist(err) {
		log.Fatal("A file or directory named `%s` already exists. Cannot continue packaging init for %s.", packagingFormat, packagingFormat)
		os.Exit(1)
	}
	err := os.MkdirAll(packagingFormatPath(packagingFormat), 0775)
	if err != nil {
		log.Fatal("Failed to create %s directory %s: %v", packagingFormat, packagingFormatPath(packagingFormat), err)
		os.Exit(1)
	}
}

func assertPackagingFormatInitialized(packagingFormat string) {
	if _, err := os.Stat(packagingFormatPath(packagingFormat)); os.IsNotExist(err) {
		log.Fatal("%s is not initialized for packaging. Please run `hover init-packaging %s` first.", packagingFormat, packagingFormat)
		os.Exit(1)
	}
}

func assertCorrectOS(packagingFormat string) {
	if runtime.GOOS != strings.Split(packagingFormat, "-")[0] {
		log.Fatal("%s only works on %s", packagingFormat, strings.Split(packagingFormat, "-")[0])
		os.Exit(1)
	}
}

func removeDashesAndUnderscores(projectName string) string {
	return strings.ReplaceAll(strings.ReplaceAll(projectName, "-", ""), "_", "")
}

func printInitFinished(packagingFormat string) {
	log.Info("go/packaging/%s has been created. You can modify the configuration files and add it to git.", packagingFormat)
	log.Info("You now can package the %s using `%s`", strings.Split(packagingFormat, "-")[0], fmt.Sprintf(au.Magenta("hover build %s").String(), packagingFormat))
}

func getTemporaryBuildDirectory(projectName string, packagingFormat string) string {
	tmpPath, err := ioutil.TempDir("", "hover-build-"+projectName+"-"+packagingFormat)
	if err != nil {
		log.Fatal("Couldn't get temporary build directory: %v", err)
		os.Exit(1)
	}
	return tmpPath
}

func initLinuxSnap(projectName string) {
	packagingFormat := "linux-snap"
	assertCorrectOS(packagingFormat)
	createPackagingFormatDirectory(packagingFormat)
	snapDirectoryPath := packagingFormatPath(packagingFormat)

	snapLocalDirectoryPath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "snap", "local"))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for snap local directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(snapLocalDirectoryPath, 0775)
	if err != nil {
		log.Fatal("Failed to create snap local directory %s: %v", snapDirectoryPath, err)
		os.Exit(1)
	}

	snapcraftFilePath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "snap", "snapcraft.yaml"))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for snapcraft.yaml file %s: %v", snapcraftFilePath, err)
		os.Exit(1)
	}

	snapcraftFile, err := os.Create(snapcraftFilePath)
	if err != nil {
		log.Fatal("Failed to create snapcraft.yaml file %s: %v", snapcraftFilePath, err)
		os.Exit(1)
	}
	snapcraftFileContent := []string{
		"name: " + removeDashesAndUnderscores(projectName),
		"base: core18",
		"version: '" + getPubSpec().Version + "'",
		"summary: " + getPubSpec().Description,
		"description: |",
		"  " + getPubSpec().Description,
		"confinement: devmode",
		"grade: devel",
		"apps:",
		"  " + removeDashesAndUnderscores(projectName) + ":",
		"    command: " + projectName,
		"    desktop: local/" + projectName + ".desktop",
		"parts:",
		"  desktop:",
		"    plugin: dump",
		"    source: snap",
		"  assets:",
		"    plugin: dump",
		"    source: assets",
		"  app:",
		"    plugin: dump",
		"    source: build",
		"    stage-packages:",
	}
	for _, dependency := range linuxPackagingDependencies {
		snapcraftFileContent = append(snapcraftFileContent, "      - "+dependency)
	}

	for _, line := range snapcraftFileContent {
		if _, err := snapcraftFile.WriteString(line + "\n"); err != nil {
			log.Fatal("Could not write snapcraft.yaml: %v", err)
			os.Exit(1)
		}
	}
	err = snapcraftFile.Close()
	if err != nil {
		log.Fatal("Could not close snapcraft.yaml: %v", err)
		os.Exit(1)
	}

	desktopFilePath, err := filepath.Abs(filepath.Join(snapLocalDirectoryPath, projectName+".desktop"))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for desktop file %s: %v", desktopFilePath, err)
		os.Exit(1)
	}
	desktopFile, err := os.Create(desktopFilePath)
	if err != nil {
		log.Fatal("Failed to create desktop file %s: %v", desktopFilePath, err)
		os.Exit(1)
	}
	desktopFileContent := []string{
		"[Desktop Entry]",
		"Encoding=UTF-8",
		"Version=" + getPubSpec().Version,
		"Type=Application",
		"Terminal=false",
		"Exec=/" + projectName,
		"Name=" + projectName,
		"Icon=/icon.png",
	}

	for _, line := range desktopFileContent {
		if _, err := desktopFile.WriteString(line + "\n"); err != nil {
			log.Fatal("Could not write %s.desktop: %v", projectName, err)
			os.Exit(1)
		}
	}
	err = desktopFile.Close()
	if err != nil {
		log.Fatal("Could not close %s.desktop: %v", projectName, err)
		os.Exit(1)
	}

	printInitFinished(packagingFormat)
}

func buildLinuxSnap(projectName string) {
	packagingFormat := "linux-snap"
	assertCorrectOS(packagingFormat)
	snapcraftBin, err := exec.LookPath("snapcraft")
	if err != nil {
		log.Fatal("Failed to lookup `snapcraft` executable. Please install snapcraft.\nhttps://tutorials.ubuntu.com/tutorial/create-your-first-snap#1")
		os.Exit(1)
	}
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	log.Info("Packaging snap in %s", tmpPath)

	err = copy.Copy(filepath.Join(buildPath, "assets"), filepath.Join(tmpPath, "assets"))
	if err != nil {
		log.Fatal("Could not copy assets folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(outputDirectoryPath("linux"), filepath.Join(tmpPath, "build"))
	if err != nil {
		log.Fatal("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		log.Fatal("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}

	cmdBuildSnap := exec.Command(snapcraftBin)
	cmdBuildSnap.Dir = tmpPath
	cmdBuildSnap.Stdout = os.Stdout
	cmdBuildSnap.Stderr = os.Stderr
	cmdBuildSnap.Stdin = os.Stdin
	err = cmdBuildSnap.Run()
	if err != nil {
		log.Fatal("Failed to package snap: %v", err)
		os.Exit(1)
	}
	outputFilePath := filepath.Join(outputDirectoryPath("linux-snap"), removeDashesAndUnderscores(projectName)+"_"+runtime.GOARCH+".snap")
	err = os.Rename(filepath.Join(tmpPath, removeDashesAndUnderscores(projectName)+"_"+getPubSpec().Version+"_"+runtime.GOARCH+".snap"), outputFilePath)
	if err != nil {
		log.Fatal("Could not move snap file: %v", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpPath)
	if err != nil {
		log.Fatal("Could not remove packaging configuration folder: %v", err)
		os.Exit(1)
	}
}

func initLinuxDeb(projectName string) {
	packagingFormat := "linux-deb"
	assertCorrectOS(packagingFormat)
	author := getPubSpec().Author
	if author == "" {
		log.Warn("Missing author field in pubspec.yaml")
		u, err := user.Current()
		if err != nil {
			log.Fatal("Couldn't get current user: %v", err)
			os.Exit(1)
		}
		author = u.Username
		log.Print("Using this username from system instead: %s", author)
	}
	createPackagingFormatDirectory(packagingFormat)
	debDirectoryPath := packagingFormatPath(packagingFormat)
	debDebianDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "DEBIAN"))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for DEBIAN directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(debDebianDirectoryPath, 0775)
	if err != nil {
		log.Fatal("Failed to create DEBIAN directory %s: %v", debDebianDirectoryPath, err)
		os.Exit(1)
	}

	binDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "usr", "bin"))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for bin directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(binDirectoryPath, 0775)
	if err != nil {
		log.Fatal("Failed to create bin directory %s: %v", binDirectoryPath, err)
		os.Exit(1)
	}
	applicationsDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "usr", "share", "applications"))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for applications directory: %v", err)
		os.Exit(1)
	}
	err = os.MkdirAll(applicationsDirectoryPath, 0775)
	if err != nil {
		log.Fatal("Failed to create applications directory %s: %v", applicationsDirectoryPath, err)
		os.Exit(1)
	}

	controlFilePath, err := filepath.Abs(filepath.Join(debDebianDirectoryPath, "control"))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for control file %s: %v", controlFilePath, err)
		os.Exit(1)
	}

	controlFile, err := os.Create(controlFilePath)
	if err != nil {
		log.Fatal("Failed to create control file %s: %v", controlFilePath, err)
		os.Exit(1)
	}
	controlFileContent := []string{
		"Package: " + removeDashesAndUnderscores(projectName),
		"Architecture: " + runtime.GOARCH,
		"Maintainer: @" + getPubSpec().Author,
		"Priority: optional",
		"Version: " + getPubSpec().Version,
		"Description: " + getPubSpec().Description,
		"Depends: " + strings.Join(linuxPackagingDependencies, ","),
	}

	for _, line := range controlFileContent {
		if _, err := controlFile.WriteString(line + "\n"); err != nil {
			log.Fatal("Could not write control file: %v", err)
			os.Exit(1)
		}
	}
	err = controlFile.Close()
	if err != nil {
		log.Fatal("Could not close control file: %v", err)
		os.Exit(1)
	}

	binFilePath, err := filepath.Abs(filepath.Join(binDirectoryPath, removeDashesAndUnderscores(projectName)))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for bin file %s: %v", binFilePath, err)
		os.Exit(1)
	}

	binFile, err := os.Create(binFilePath)
	if err != nil {
		log.Fatal("Failed to create bin file %s: %v", controlFilePath, err)
		os.Exit(1)
	}
	binFileContent := []string{
		"#!/bin/sh",
		"/usr/lib/" + projectName + "/" + projectName,
	}
	for _, line := range binFileContent {
		if _, err := binFile.WriteString(line + "\n"); err != nil {
			log.Fatal("Could not write bin file: %v", err)
			os.Exit(1)
		}
	}
	err = binFile.Close()
	if err != nil {
		log.Fatal("Could not close bin file: %v", err)
		os.Exit(1)
	}
	err = os.Chmod(binFilePath, 0777)
	if err != nil {
		log.Fatal("Failed to change file permissions for bin file: %v", err)
		os.Exit(1)
	}

	desktopFilePath, err := filepath.Abs(filepath.Join(applicationsDirectoryPath, projectName+".desktop"))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for desktop file %s: %v", desktopFilePath, err)
		os.Exit(1)
	}
	desktopFile, err := os.Create(desktopFilePath)
	if err != nil {
		log.Fatal("Failed to create desktop file %s: %v", desktopFilePath, err)
		os.Exit(1)
	}
	desktopFileContent := []string{
		"[Desktop Entry]",
		"Encoding=UTF-8",
		"Version=" + getPubSpec().Version,
		"Type=Application",
		"Terminal=false",
		"Exec=/usr/bin/" + projectName,
		"Name=" + projectName,
		"Icon=/usr/lib/" + projectName + "/assets/icon.png",
	}
	for _, line := range desktopFileContent {
		if _, err := desktopFile.WriteString(line + "\n"); err != nil {
			log.Fatal("Could not write %s.desktop file: %v", projectName, err)
			os.Exit(1)
		}
	}
	err = desktopFile.Close()
	if err != nil {
		log.Fatal("Could not close %s.desktop file: %v", projectName, err)
		os.Exit(1)
	}

	printInitFinished(packagingFormat)
}

func buildLinuxDeb(projectName string) {
	packagingFormat := "linux-deb"
	assertCorrectOS(packagingFormat)
	dpkgDebBin, err := exec.LookPath("dpkg-deb")
	if err != nil {
		log.Fatal("Failed to lookup `dpkg-deb` executable. Please install dpkg-deb.")
		os.Exit(1)
	}
	tmpPath := getTemporaryBuildDirectory(projectName, packagingFormat)
	log.Info("Packaging deb in %s", tmpPath)

	libDirectoryPath, err := filepath.Abs(filepath.Join(tmpPath, "usr", "lib"))
	if err != nil {
		log.Fatal("Failed to resolve absolute path for bin directory: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(outputDirectoryPath("linux"), filepath.Join(libDirectoryPath, projectName))
	if err != nil {
		log.Fatal("Could not copy build folder: %v", err)
		os.Exit(1)
	}
	err = copy.Copy(packagingFormatPath(packagingFormat), filepath.Join(tmpPath))
	if err != nil {
		log.Fatal("Could not copy packaging configuration folder: %v", err)
		os.Exit(1)
	}
	outputFileName := removeDashesAndUnderscores(projectName) + "_" + runtime.GOARCH + ".deb"
	outputFilePath := filepath.Join(outputDirectoryPath("linux-deb"), outputFileName)

	cmdBuildDeb := exec.Command(dpkgDebBin, "--build", ".", outputFileName)
	cmdBuildDeb.Dir = tmpPath
	cmdBuildDeb.Stdout = os.Stdout
	cmdBuildDeb.Stderr = os.Stderr
	cmdBuildDeb.Stdin = os.Stdin
	err = cmdBuildDeb.Run()
	if err != nil {
		log.Fatal("Failed to package deb: %v", err)
		os.Exit(1)
	}
	err = os.Rename(filepath.Join(tmpPath, outputFileName), outputFilePath)
	if err != nil {
		log.Fatal("Could not move deb file: %v", err)
		os.Exit(1)
	}
	err = os.RemoveAll(tmpPath)
	if err != nil {
		log.Fatal("Could not remove packaging configuration folder: %v", err)
		os.Exit(1)
	}
}
