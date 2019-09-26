package packaging

import (
	"fmt"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-flutter-desktop/hover/internal/build"
)

var packagingPath = filepath.Join(build.BuildPath, "packaging")

func packagingFormatPath(packagingFormat string) string {
	directoryPath, err := filepath.Abs(filepath.Join(packagingPath, packagingFormat))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for %s directory: %v\n", packagingFormat, err)
		os.Exit(1)
	}
	return directoryPath
}

func createPackagingFormatDirectory(packagingFormat string) {
	if _, err := os.Stat(packagingFormatPath(packagingFormat)); !os.IsNotExist(err) {
		fmt.Printf("hover: A file or directory named `%s` already exists. Cannot continue packaging init for %s.\n", packagingFormat, packagingFormat)
		os.Exit(1)
	}
	err := os.MkdirAll(packagingFormatPath(packagingFormat), 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create %s directory %s: %v\n", packagingFormat, packagingFormatPath(packagingFormat), err)
		os.Exit(1)
	}
}

func AssertPackagingFormatInitialized(packagingFormat string) {
	if _, err := os.Stat(packagingFormatPath(packagingFormat)); os.IsNotExist(err) {
		fmt.Printf("hover: %s is not initialized for packaging. Please run `hover init-packaging %s` first.\n", packagingFormat, packagingFormat)
		os.Exit(1)
	}
}

func removeDashesAndUnderscores(projectName string) string {
	return strings.ReplaceAll(strings.ReplaceAll(projectName, "-", ""), "_", "")
}

func printInitFinished(packagingFormat string) {
	fmt.Printf("hover: go/packaging/%s has been created. You can modify the configuration files and add it to git.\n", packagingFormat)
	fmt.Printf("hover: You now can package the %s using `hover build %s`\n", strings.Split(packagingFormat, "-")[0], packagingFormat)
}

func printPackagingFinished(packagingFormat string) {
	fmt.Printf("hover: Successfully packaged %s\n", strings.Split(packagingFormat, "-")[1])
}

func getTemporaryBuildDirectory(projectName string, packagingFormat string) string {
	tmpPath, err := ioutil.TempDir("", "hover-build-"+projectName+"-"+packagingFormat)
	if err != nil {
		fmt.Printf("hover: Couldn't get temporary build directory: %v\n", err)
		os.Exit(1)
	}
	return tmpPath
}

func DockerInstalled() bool {
	if build.DockerBin == "" {
		fmt.Println("hover: To use packaging, Docker needs to be installed.\nhttps://docs.docker.com/install")
	}
	return build.DockerBin != ""
}

func getAuthor() string {
	author := pubspec.GetPubSpec().Author
	if author == "" {
		fmt.Println("hover: Missing author field in pubspec.yaml")
		u, err := user.Current()
		if err != nil {
			fmt.Printf("hover: Couldn't get current user: %v\n", err)
			os.Exit(1)
		}
		author = u.Username
		fmt.Printf("hover: Using this username from system instead: %s\n", author)
	}
	return author
}

func createDockerfile(packagingFormat string) {
	dockerFilePath, err := filepath.Abs(filepath.Join(packagingFormatPath(packagingFormat), "Dockerfile"))
	if err != nil {
		fmt.Printf("hover: Failed to resolve absolute path for Dockerfile %s: %v\n", dockerFilePath, err)
		os.Exit(1)
	}
	dockerFile, err := os.Create(dockerFilePath)
	if err != nil {
		fmt.Printf("hover: Failed to create Dockerfile %s: %v\n", dockerFilePath, err)
		os.Exit(1)
	}
	dockerFileContent := []string{}
	if packagingFormat == "linux-snap" {
		dockerFileContent = []string{
			"FROM snapcore/snapcraft",
		}
	} else if packagingFormat == "linux-deb" {
		dockerFileContent = []string{
			"FROM ubuntu:bionic",
		}
	} else if packagingFormat == "windows-msi" {
		dockerFileContent = []string{
			"FROM ubuntu:bionic",
			"RUN apt-get update && apt-get install wixl -y",
		}
	} else {
		fmt.Printf("hover: Tried to create Dockerfile for unknown packaging format %s\n", packagingFormat)
		os.Exit(1)
	}

	for _, line := range dockerFileContent {
		if _, err := dockerFile.WriteString(line + "\n"); err != nil {
			fmt.Printf("hover: Could not write Dockerfile: %v\n", err)
			os.Exit(1)
		}
	}
	err = dockerFile.Close()
	if err != nil {
		fmt.Printf("hover: Could not close Dockerfile: %v\n", err)
		os.Exit(1)
	}
}

func runDockerPackaging(path string, packagingFormat string, command []string) {
	dockerBuildCmd := exec.Command(build.DockerBin, "build", "-t", "hover-build-packaging-"+packagingFormat, ".")
	dockerBuildCmd.Stdout = os.Stdout
	dockerBuildCmd.Stderr = os.Stderr
	dockerBuildCmd.Dir = packagingFormatPath(packagingFormat)
	err := dockerBuildCmd.Run()
	if err != nil {
		fmt.Printf("hover: Docker build failed: %v\n", err)
		os.Exit(1)
	}
	u, err := user.Current()
	if err != nil {
		fmt.Printf("hover: Couldn't get current user: %v\n", err)
		os.Exit(1)
	}
	args := []string{
		"run",
		"-w", "/app",
		"-v", path + ":/app",
	}
	args = append(args, "hover-build-packaging-"+packagingFormat)
	chownStr := ""
	if runtime.GOOS != "windows" {
		chownStr = fmt.Sprintf(" && chown %s:%s * -R", u.Uid, u.Gid)
	}
	args = append(args, "bash", "-c", fmt.Sprintf("%s%s", strings.Join(command, " "), chownStr))
	dockerRunCmd := exec.Command(build.DockerBin, args...)
	dockerRunCmd.Stderr = os.Stderr
	dockerRunCmd.Stdout = os.Stdout
	dockerRunCmd.Dir = path
	err = dockerRunCmd.Run()
	if err != nil {
		fmt.Printf("hover: Docker run failed: %v\n", err)
		os.Exit(1)
	}
}
