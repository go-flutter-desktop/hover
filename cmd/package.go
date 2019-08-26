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

var install bool

func init() {
	rootCmd.AddCommand(packageCmd)
	packageSnapCmd.Flags().BoolVarP(&install, "install", "i", false, "Install the snap after packaging it")
	packageCmd.AddCommand(packageSnapCmd)
	packageDebCmd.Flags().BoolVarP(&install, "install", "i", false, "Install the deb after packaging it")
	packageCmd.AddCommand(packageDebCmd)
}

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Package a desktop release",
}

var linuxPackageDependencies = []string{"libx11-6", "libxrandr2", "libxcursor1", "libxinerama1"}

var packageSnapCmd = &cobra.Command{
	Use:   "snap",
	Short: "Package a desktop release as snap",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := assertInFlutterProject().Name
		assertHoverInitialized()

		// Hardcode target to the current OS (no cross-compile support yet)
		targetOS := runtime.GOOS

		packageSnap(projectName, targetOS)
	},
}

var packageDebCmd = &cobra.Command{
	Use:   "deb",
	Short: "Package a desktop release as deb",
	Run: func(cmd *cobra.Command, args []string) {
		projectName := assertInFlutterProject().Name
		assertHoverInitialized()

		// Hardcode target to the current OS (no cross-compile support yet)
		targetOS := runtime.GOOS

		packageDeb(projectName, targetOS)
	},
}

func assertBuildOutputAvailable(projectName string, targetOS string) {
	if _, err := os.Stat(outputBinaryPath(projectName, targetOS)); os.IsNotExist(err) {
		fmt.Println("hover: No build outputs found\nhover: Please run `hover build` first")
		os.Exit(1)
	}
}

func packageSnap(projectName string, targetOS string) {
	if targetOS != "linux" {
		fmt.Println("hover: Snap only works on linux")
		os.Exit(1)
	}
	assertBuildOutputAvailable(projectName, targetOS)
	var snapcraftBin, err = exec.LookPath("snapcraft")
	if err != nil {
		fmt.Println("hover: Failed to lookup `snapcraft` executable. Please install snapcraft.\nhttps://tutorials.ubuntu.com/tutorial/create-your-first-snap#1")
		os.Exit(1)
	}
	copyAsset("app/gitignore", filepath.Join("desktop", ".gitignore"))
	snapDirectoryPath, err := filepath.Abs(filepath.Join("desktop", "snap"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for snap directory: %v\n", err)
		os.Exit(1)
	}
	err = os.RemoveAll(snapDirectoryPath)
	if err != nil {
		fmt.Printf("hover: Failed to clean snap directory %s: %v\n", snapDirectoryPath, err)
		os.Exit(1)
	}
	err = os.MkdirAll(snapDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create snap directory %s: %v\n", snapDirectoryPath, err)
		os.Exit(1)
	}

	snapLocalDirectoryPath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "local"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for snap local directory: %v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(snapLocalDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create snap local directory %s: %v\n", snapDirectoryPath, err)
		os.Exit(1)
	}

	snapcraftFilePath, err := filepath.Abs(filepath.Join(snapDirectoryPath, "snapcraft.yaml"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for snapcraft.yaml file %s: %v\n", snapcraftFilePath, err)
		os.Exit(1)
	}

	snapcraftFile, err := os.Create(snapcraftFilePath)
	if err != nil {
		fmt.Printf("hover: Failed to create snapcraft.yaml file %s: %v\n", snapcraftFilePath, err)
		os.Exit(1)
	}
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
	for _, dependency := range linuxPackageDependencies {
		snapcraftFile.WriteString("      - " + dependency + "\n")
	}
	snapcraftFile.WriteString("  desktop:\n")
	snapcraftFile.WriteString("    plugin: dump\n")
	snapcraftFile.WriteString("    source: snap\n")
	snapcraftFile.WriteString("  assets:\n")
	snapcraftFile.WriteString("    plugin: dump\n")
	snapcraftFile.WriteString("    source: assets\n")
	snapcraftFile.WriteString("apps:\n")
	snapcraftFile.WriteString("  " + projectName + ":\n")
	snapcraftFile.WriteString("    command: " + projectName + "\n")
	snapcraftFile.WriteString("    desktop: local/" + projectName + ".desktop\n")
	snapcraftFile.Close()

	splitProjectName := strings.Split(projectName, "")
	splitProjectName[0] = strings.ToUpper(splitProjectName[0])
	titledProjectName := strings.Join(splitProjectName, "")
	desktopFilePath, err := filepath.Abs(filepath.Join(snapLocalDirectoryPath, projectName+".desktop"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for desktop file %s: %v\n", desktopFilePath, err)
		os.Exit(1)
	}
	desktopFile, err := os.Create(desktopFilePath)
	if err != nil {
		fmt.Printf("hover: Failed to create desktop file %s: %v\n", desktopFilePath, err)
		os.Exit(1)
	}
	desktopFile.WriteString("[Desktop Entry]\n")
	desktopFile.WriteString("Encoding=UTF-8\n")
	desktopFile.WriteString("Version=" + assertInFlutterProject().Version + "\n")
	desktopFile.WriteString("Type=Application\n")
	desktopFile.WriteString("Terminal=false\n")
	desktopFile.WriteString("Exec=/" + projectName + "\n")
	desktopFile.WriteString("Name=" + titledProjectName + "\n")
	desktopFile.WriteString("Icon=/icon.png\n")
	desktopFile.Close()

	fmt.Println("hover: Packaging snap...")
	cmdBuildSnap := exec.Command(snapcraftBin)
	cmdBuildSnap.Dir = filepath.Join(snapDirectoryPath, "..")
	cmdBuildSnap.Stdout = os.Stdout
	cmdBuildSnap.Stderr = os.Stderr
	cmdBuildSnap.Stdin = os.Stdin
	cmdBuildSnap.Start()
	cmdBuildSnap.Wait()

	outputFilePath := filepath.Join(outputDirectoryPath(targetOS), projectName+"_"+runtime.GOARCH+".snap")
	os.Rename(filepath.Join(snapDirectoryPath, "..", projectName+"_"+assertInFlutterProject().Version+"_"+runtime.GOARCH+".snap"), outputFilePath)
	if install {
		fmt.Println("hover: Installing snap...")
		var snapBin, err = exec.LookPath("snap")
		if err != nil {
			fmt.Println("hover: Failed to lookup `snap` executable. Please install snap.")
			os.Exit(1)
		}
		cmdInstallSnap := exec.Command("/bin/sh", "-c", "sudo "+snapBin+" install "+outputFilePath+" --devmode")
		cmdInstallSnap.Stdout = os.Stdout
		cmdInstallSnap.Stderr = os.Stderr
		cmdInstallSnap.Stdin = os.Stdin
		cmdInstallSnap.Start()
		cmdInstallSnap.Wait()
	}
}

func packageDeb(projectName string, targetOS string) {
	if targetOS != "linux" {
		fmt.Println("hover: Deb only works on linux")
		os.Exit(1)
	}
	assertBuildOutputAvailable(projectName, targetOS)
	var dpkgDebBin, err = exec.LookPath("dpkg-deb")
	if err != nil {
		fmt.Println("hover: Failed to lookup `dpkg-deb` executable. Please install dpkg-deb.")
		os.Exit(1)
	}
	copyAsset("app/gitignore", filepath.Join("desktop", ".gitignore"))
	debDirectoryPath, err := filepath.Abs(filepath.Join("desktop", "deb"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for deb directory: %v\n", err)
		os.Exit(1)
	}
	err = os.RemoveAll(debDirectoryPath)
	if err != nil {
		fmt.Printf("hover: Failed to clean deb directory %s: %v\n", debDirectoryPath, err)
		os.Exit(1)
	}
	err = os.MkdirAll(debDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create deb directory %s: %v\n", debDirectoryPath, err)
		os.Exit(1)
	}

	debDebianDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "DEBIAN"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for DEBIAN directory: %v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(debDebianDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create DEBIAN directory %s: %v\n", debDebianDirectoryPath, err)
		os.Exit(1)
	}

	binDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "usr", "bin"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for bin directory: %v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(binDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create bin directory %s: %v\n", binDirectoryPath, err)
		os.Exit(1)
	}
	applicationsDirectoryPath, err := filepath.Abs(filepath.Join(debDirectoryPath, "usr", "share", "applications"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for applications directory: %v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(applicationsDirectoryPath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create applications directory %s: %v\n", applicationsDirectoryPath, err)
		os.Exit(1)
	}

	controlFilePath, err := filepath.Abs(filepath.Join(debDebianDirectoryPath, "control"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for control file %s: %v\n", controlFilePath, err)
		os.Exit(1)
	}

	controlFile, err := os.Create(controlFilePath)
	if err != nil {
		fmt.Printf("hover: Failed to create control file %s: %v\n", controlFilePath, err)
		os.Exit(1)
	}
	controlFile.WriteString("Package: " + projectName + "\n")
	controlFile.WriteString("Architecture: " + runtime.GOARCH + "\n")
	controlFile.WriteString("Maintainer: @" + assertInFlutterProject().Author + "\n")
	controlFile.WriteString("Priority: optional\n")
	controlFile.WriteString("Version: " + assertInFlutterProject().Version + "\n")
	controlFile.WriteString("Description: " + assertInFlutterProject().Description + "\n")
	controlFile.WriteString("Depends: " + strings.Join(linuxPackageDependencies, ",") + "\n")
	controlFile.Close()

	binFilePath, err := filepath.Abs(filepath.Join(binDirectoryPath, "ginko"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for bin file %s: %v\n", binFilePath, err)
		os.Exit(1)
	}

	binFile, err := os.Create(binFilePath)
	if err != nil {
		fmt.Printf("hover: Failed to create bin file %s: %v\n", controlFilePath, err)
		os.Exit(1)
	}
	binFile.WriteString("#!/bin/sh\n")
	binFile.WriteString("/usr/bin/" + projectName + "files/" + projectName + "\n")
	binFile.Close()
	os.Chmod(binFilePath, 0777)

	splitProjectName := strings.Split(projectName, "")
	splitProjectName[0] = strings.ToUpper(splitProjectName[0])
	titledProjectName := strings.Join(splitProjectName, "")
	desktopFilePath, err := filepath.Abs(filepath.Join(applicationsDirectoryPath, projectName+".desktop"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for desktop file %s: %v\n", desktopFilePath, err)
		os.Exit(1)
	}
	desktopFile, err := os.Create(desktopFilePath)
	if err != nil {
		fmt.Printf("hover: Failed to create desktop file %s: %v\n", desktopFilePath, err)
		os.Exit(1)
	}
	desktopFile.WriteString("[Desktop Entry]\n")
	desktopFile.WriteString("Encoding=UTF-8\n")
	desktopFile.WriteString("Version=" + assertInFlutterProject().Version + "\n")
	desktopFile.WriteString("Type=Application\n")
	desktopFile.WriteString("Terminal=false\n")
	desktopFile.WriteString("Exec=/usr/bin/" + projectName + "\n")
	desktopFile.WriteString("Name=" + titledProjectName + "\n")
	desktopFile.WriteString("Icon=/usr/bin/" + projectName + "files/assets/icon.png\n")
	desktopFile.Close()

	copy.Copy(outputDirectoryPath(targetOS), filepath.Join(binDirectoryPath, projectName+"files"))

	fmt.Println("hover: Packaging deb...")
	cmdBuildDeb := exec.Command(dpkgDebBin, "--build", "deb")
	cmdBuildDeb.Dir = filepath.Join(debDirectoryPath, "..")
	cmdBuildDeb.Stdout = os.Stdout
	cmdBuildDeb.Stderr = os.Stderr
	cmdBuildDeb.Stdin = os.Stdin
	cmdBuildDeb.Start()
	cmdBuildDeb.Wait()

	outputFilePath := filepath.Join(outputDirectoryPath(targetOS), projectName+"_"+runtime.GOARCH+".deb")
	os.Rename(filepath.Join(debDirectoryPath, "..", "deb.deb"), outputFilePath)
	if install {
		fmt.Println("hover: Installing deb...")
		var dpkgBin, err = exec.LookPath("dpkg")
		if err != nil {
			fmt.Println("hover: Failed to lookup `dpkg` executable. Please install dpkg.")
			os.Exit(1)
		}
		cmdInstallDeb := exec.Command("/bin/sh", "-c", "sudo "+dpkgBin+" -i "+outputFilePath)
		cmdInstallDeb.Stdout = os.Stdout
		cmdInstallDeb.Stderr = os.Stderr
		cmdInstallDeb.Stdin = os.Stdin
		cmdInstallDeb.Start()
		cmdInstallDeb.Wait()
	}
}
