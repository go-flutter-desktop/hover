package cmd

import (
	"fmt"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func init() {
	rootCmd.AddCommand(packageCmd)
	packageCmd.AddCommand(packageSnapcraftCmd)
}

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Package a desktop release",
}

var packageSnapcraftCmd = &cobra.Command{
	Use:   "snapcraft",
	Short: "Package a desktop release as snap",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := assertInFlutterProject().Name
		assertHoverInitialized()

		// Hardcode target to the current OS (no cross-compile support yet)
		targetOS := runtime.GOOS

		packageSnapcraft(projectName, targetOS, nil)
	},
}

func packageSnapcraft(projectName string, targetOS string, vmArguments []string) {
	if targetOS != "linux" {
		fmt.Println("hover: snapcraft only works on linux")
		os.Exit(1)
	}
	if _, err := os.Stat(outputBinaryPath(projectName, targetOS)); os.IsNotExist(err) {
		fmt.Println("hover: no build outputs found\nhover: Please run `hover build` first")
		os.Exit(1)
	}
	var snapcraftBin, err = exec.LookPath("snapcraft")
	if err != nil {
		fmt.Println("hover: Failed to lookup `snapcraft` executable. Please install snapcraft.\nhttps://tutorials.ubuntu.com/tutorial/create-your-first-snap#1")
		os.Exit(1)
	}
	snapDirectoryPath, err := filepath.Abs(filepath.Join("desktop", "snap"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for snap directory: %v\n", err)
		os.Exit(1)
	}
	err = os.RemoveAll(snapDirectoryPath)
	if err != nil {
		fmt.Printf("hover: failed to clean snap directory %s: %v\n", snapDirectoryPath, err)
		os.Exit(1)
	}
	err = os.MkdirAll(snapDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: failed to create snap directory %s: %v\n", snapDirectoryPath, err)
		os.Exit(1)
	}

	snapLocalDirectoryPath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "local"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for snap local directory: %v\n", err)
		os.Exit(1)
	}
	err = os.RemoveAll(snapLocalDirectoryPath)
	if err != nil {
		fmt.Printf("hover: failed to clean snap local directory %s: %v\n", snapDirectoryPath, err)
		os.Exit(1)
	}
	err = os.MkdirAll(snapLocalDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: failed to create snap local directory %s: %v\n", snapDirectoryPath, err)
		os.Exit(1)
	}

	gitignoreFilePath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "..", ".gitignore"))
	if err != nil {
		fmt.Printf("hover: failed to create .gitignore file %s: %v\n", gitignoreFilePath, err)
		os.Exit(1)
	}

	gitignoreFile, err := os.Create(gitignoreFilePath)
	gitignoreFile.WriteString("build/\n")
	gitignoreFile.WriteString("snap/\n")
	gitignoreFile.Close()

	snapcraftFilePath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "snapcraft.yaml"))
	if err != nil {
		fmt.Printf("hover: failed to create snapcraft.yaml file %s: %v\n", snapcraftFilePath, err)
		os.Exit(1)
	}

	snapcraftFile, err := os.Create(snapcraftFilePath)
	snapcraftFile.WriteString("name: " + projectName + "\n")
	snapcraftFile.WriteString("base: core18\n")
	snapcraftFile.WriteString("version: '" + assertInFlutterProject().Version + "'\n")
	snapcraftFile.WriteString("summary: " + assertInFlutterProject().Description + "\n")
	snapcraftFile.WriteString("description: |\n")
	snapcraftFile.WriteString("  " + assertInFlutterProject().Description + "\n")
	snapcraftFile.WriteString("confinement: strict\n")
	snapcraftFile.WriteString("grade: stable\n")
	snapcraftFile.WriteString("parts:\n")
	snapcraftFile.WriteString("  app:\n")
	snapcraftFile.WriteString("    plugin: dump\n")
	snapcraftFile.WriteString("    source: build/outputs/linux\n")
	snapcraftFile.WriteString("    stage-packages:\n")
	snapcraftFile.WriteString("      - libx11-6\n")
	snapcraftFile.WriteString("      - libxrandr2\n")
	snapcraftFile.WriteString("      - libxcursor1\n")
	snapcraftFile.WriteString("      - libxinerama1\n")
	snapcraftFile.WriteString("  desktop:\n")
	snapcraftFile.WriteString("    plugin: dump\n")
	snapcraftFile.WriteString("    source: snap\n")
	snapcraftFile.WriteString("apps:\n")
	snapcraftFile.WriteString("  " + projectName + ":\n")
	snapcraftFile.WriteString("    command: " + projectName + "\n")
	snapcraftFile.WriteString("    desktop: local/" + projectName + ".desktop\n")
	snapcraftFile.Close()

	assetIconFilePath := filepath.Join(snapDirectoryPath, "..", "assets", "icon.png")
	desktopIconFilePath := filepath.Join(snapLocalDirectoryPath, "icon.png")
	copy.Copy(assetIconFilePath, desktopIconFilePath)
	splitProjectName := strings.Split(projectName, "")
	splitProjectName[0] = strings.ToUpper(splitProjectName[0])
	titledProjectName := strings.Join(splitProjectName, "")
	desktopFilePath, err := filepath.Abs(filepath.Join(snapLocalDirectoryPath, projectName+".desktop"))
	if err != nil {
		fmt.Printf("hover: failed to create "+projectName+".desktop file %s: %v\n", desktopFilePath, err)
		os.Exit(1)
	}
	desktopFile, err := os.Create(desktopFilePath)
	desktopFile.WriteString("[Desktop Entry]\n")
	desktopFile.WriteString("Encoding=UTF-8\n")
	desktopFile.WriteString("Version=" + assertInFlutterProject().Version + "\n")
	desktopFile.WriteString("Type=Application\n")
	desktopFile.WriteString("Terminal=false\n")
	desktopFile.WriteString("Exec=/" + projectName + "\n")
	desktopFile.WriteString("Name=" + titledProjectName + "\n")
	desktopFile.WriteString("Icon=/local/icon.png\n")
	desktopFile.Close()

	cmdBuildSnap := exec.Command(snapcraftBin)
	cmdBuildSnap.Dir = filepath.Join(snapDirectoryPath, "..")
	cmdBuildSnap.Stdout = os.Stdout
	cmdBuildSnap.Stderr = os.Stderr
	cmdBuildSnap.Stdin = os.Stdin
	cmdBuildSnap.Start()
	cmdBuildSnap.Wait()

	snapFilePath := projectName + "_" + assertInFlutterProject().Version + "_" + runtime.GOARCH + ".snap"

	os.Rename(filepath.Join(snapDirectoryPath, "..", snapFilePath), filepath.Join(outputDirectoryPath(targetOS), snapFilePath))
}
