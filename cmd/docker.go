package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-flutter-desktop/hover/cmd/packaging"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/logstreamer"
)

func dockerHoverBuild(targetOS string, packagingTask packaging.Task, buildFlags []string, vmArguments []string) {
	var err error
	dockerBin := build.DockerBin()

	hoverCacheDir := filepath.Join(buildOrRunCachePath, "hover")

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
	goproxy, err := exec.Command("go", "env", "GOPROXY").Output()
	if err != nil {
		log.Errorf("Failed to get GOPROXY: %v", err)
	}
	if string(goproxy) != "" {
		dockerArgs = append(dockerArgs, "--env", "GOPROXY="+string(goproxy))
	}
	goprivate, err := exec.Command("go", "env", "GOPRIVATE").Output()
	if err != nil {
		log.Errorf("Failed to get GOPRIVATE: %v", err)
	}
	if string(goprivate) != "" {
		dockerArgs = append(dockerArgs, "--env", "GOPRIVATE="+string(goprivate))
	}
	if len(vmArguments) > 0 {
		// I (GeertJohan) am not too happy with this, it make the hover inside
		// the container aware of it being inside the container. But for now
		// this is the best way to go about.
		//
		// HOVER_BUILD_INDOCKER_VMARGS is explicitly not document, it is not
		// intended to be abused and may disappear at any time.
		dockerArgs = append(dockerArgs, "--env", "HOVER_IN_DOCKER_BUILD_VMARGS="+strings.Join(vmArguments, ","))
	}

	version := hoverVersion()
	if version == "(devel)" {
		version = "latest"
	}
	dockerImage := "goflutter/hover:" + version
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
