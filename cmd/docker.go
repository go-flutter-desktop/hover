package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/go-flutter-desktop/hover/cmd/packaging"
	"github.com/go-flutter-desktop/hover/internal/config"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/logstreamer"
)

func dockerHoverBuild(targetOS string, packagingTask packaging.Task, buildFlags []string, vmArguments []string) {
	if buildCachePath == "" && config.GetConfig().CachePath != "" {
		buildCachePath = config.GetConfig().CachePath
	}
	dockerBin, err := exec.LookPath("docker")
	if err != nil {
		log.Printf("Docker is not installed: %v", err)
		os.Exit(1)
	}

	if buildCachePath == "" {
		buildCachePath, err = os.UserCacheDir()
		if err != nil {
			log.Errorf("Cannot get the path for the user cache directory: %v", err)
			os.Exit(1)
		}
	}
	hoverCacheDir := filepath.Join(buildCachePath, "hover")

	engineCacheDir := filepath.Join(hoverCacheDir, "engine")
	err = os.MkdirAll(engineCacheDir, 0755)
	if err != nil {
		log.Errorf("Cannot create the engine cache path in the user cache directory: %v", err)
		os.Exit(1)
	}

	dockerGoCacheDir := filepath.Join(hoverCacheDir, "docker-go-cache")
	err = os.MkdirAll(dockerGoCacheDir, 0755)
	if err != nil {
		log.Errorf("Cannot create the docker-go-cache path in the user cache directory: %v", err)
		os.Exit(1)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Errorf("Cannot get the path for current directory %s", err)
		os.Exit(1)
	}
	log.Infof("Compiling go binary using docker container")

	dockerArgs := []string{
		"run",
		"--rm",
		"--mount", "type=bind,source=" + wd + ",target=/app",
		"--mount", "type=bind,source=" + engineCacheDir + ",target=/root/.cache/hover/engine",
		"--mount", "type=bind,source=" + dockerGoCacheDir + ",target=/go-cache",
		"--env", "GOCACHE=/go-cache",
	}
	if runtime.GOOS != "windows" {
		currentUser, err := user.Current()
		if err != nil {
			log.Errorf("Couldn't get current user info: %v", err)
			os.Exit(1)
		}
		dockerArgs = append(dockerArgs, "--env", "HOVER_SAFE_CHOWN_UID="+currentUser.Uid)
		dockerArgs = append(dockerArgs, "--env", "HOVER_SAFE_CHOWN_GID="+currentUser.Gid)
	}
	if goproxy := os.Getenv("GOPROXY"); goproxy != "" {
		dockerArgs = append(dockerArgs, "--env", "GOPROXY="+goproxy)
	}
	if goprivate := os.Getenv("GOPRIVATE"); goprivate != "" {
		dockerArgs = append(dockerArgs, "--env", "GOPRIVATE="+goprivate)
	}
	// TODO: Use hover container of version of the current running hover.
	// 		Use debug package to obtain Module info which contains the version.
	dockerImage := "goflutter/hover:latest"
	dockerArgs = append(dockerArgs, dockerImage)
	targetOSAndPackaging := targetOS
	if packName := packagingTask.Name(); packName != "" {
		targetOSAndPackaging += "-" + packName
	}
	hoverCommand := []string{"hover-safe.sh", "build", targetOSAndPackaging}
	hoverCommand = append(hoverCommand, buildFlags...)
	dockerArgs = append(dockerArgs, hoverCommand...)

	dockerRunCmd := exec.Command(dockerBin, dockerArgs...)
	// TODO: remove debug line
	fmt.Printf("Running this docker command: %v\n", dockerRunCmd.String())
	dockerRunCmd.Stderr = logstreamer.NewLogstreamerForStderr("docker container: ")
	dockerRunCmd.Stdout = logstreamer.NewLogstreamerForStdout("docker container: ")
	dockerRunCmd.Dir = wd
	err = dockerRunCmd.Run()
	if err != nil {
		log.Errorf("Docker run failed: %v", err)
		os.Exit(1)
	}
	log.Infof("Docker run completed")
}
