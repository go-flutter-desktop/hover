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
	"github.com/go-flutter-desktop/hover/internal/execx"
	"github.com/go-flutter-desktop/hover/internal/logstreamer"
	"github.com/go-flutter-desktop/hover/internal/logx"
	"github.com/go-flutter-desktop/hover/internal/tracex"
)

func dockerHoverBuild(targetOS string, packagingTask packaging.Task, buildFlags []string, vmArguments []string) {
	initBuildParameters(targetOS)
	var err error
	dockerBin := build.DockerBin()

	hoverCacheDir := filepath.Join(buildCachePath, "hover")

	engineCacheDir := filepath.Join(hoverCacheDir, "engine")
	err = os.MkdirAll(engineCacheDir, 0755)
	if err != nil {
		logx.Errorf("Cannot create the engine cache path in the user cache directory: %v", err)
		os.Exit(1)
	}

	dockerGoCacheDir := filepath.Join(hoverCacheDir, "docker-go-cache")
	err = os.MkdirAll(dockerGoCacheDir, 0755)
	if err != nil {
		logx.Errorf("Cannot create the docker-go-cache path in the user cache directory: %v", err)
		os.Exit(1)
	}

	wd, err := os.Getwd()
	if err != nil {
		logx.Errorf("Cannot get the path for current directory %s", err)
		os.Exit(1)
	}
	logx.Infof("Compiling go binary using docker container")

	dockerArgs := []string{
		"run",
		"--rm",
		"--mount", "type=bind,source=" + wd + ",target=/app",
		"--mount", "type=bind,source=" + engineCacheDir + ",target=/root/.cache/hover/engine",
		"--mount", "type=bind,source=" + dockerGoCacheDir + ",target=/go-cache",
		"--env", "GOCACHE=/go-cache",
	}

	// override the docker image hover binary builtin with the local version.
	if path, err := os.Executable(); err == nil {
		dockerArgs = append(dockerArgs, "--mount", fmt.Sprintf("type=bind,source=%s,target=/go/bin/hover", path))
	} else {
		logx.Warnf("unable to located hover executable: %s", err)
	}

	if runtime.GOOS != "windows" {
		currentUser, err := user.Current()
		if err != nil {
			logx.Errorf("Couldn't get current user info: %v", err)
			os.Exit(1)
		}
		dockerArgs = append(dockerArgs, "--env", "HOVER_SAFE_CHOWN_UID="+currentUser.Uid)
		dockerArgs = append(dockerArgs, "--env", "HOVER_SAFE_CHOWN_GID="+currentUser.Gid)
	}

	if goproxy := os.Getenv("GOPROXY"); goproxy != "" {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("GOPROXY=%s", goproxy))
	}

	if goprivate := os.Getenv("GOPRIVATE"); goprivate != "" {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("GOPRIVATE=%s", goprivate))
	}

	if len(vmArguments) > 0 {
		// I (GeertJohan) am not too happy with this, it make the hover inside
		// the container aware of it being inside the container. But for now
		// this is the best way to go about.
		//
		// HOVER_BUILD_INDOCKER_VMARGS is explicitly not documented, it is not
		// intended to be abused and may disappear at any time.
		dockerArgs = append(dockerArgs, "--env", "HOVER_IN_DOCKER_BUILD_VMARGS="+strings.Join(vmArguments, ","))
	}

	dockerArgs = append(dockerArgs, buildDockerImage)
	targetOSAndPackaging := targetOS
	if packName := packagingTask.Name(); packName != "" {
		targetOSAndPackaging += "-" + packName
	}
	hoverCommand := []string{"hover-safe.sh", "build", targetOSAndPackaging}
	hoverCommand = append(hoverCommand, buildFlags...)
	dockerArgs = append(dockerArgs, hoverCommand...)

	dockerRunCmd := exec.Command(dockerBin, dockerArgs...)
	dockerRunCmd.Stderr = logstreamer.NewLogstreamerForStderr("docker container: ")
	dockerRunCmd.Stdout = logstreamer.NewLogstreamerForStdout("docker container: ")
	dockerRunCmd.Dir = wd

	tracex.Println("docker command:", execx.Describe(dockerRunCmd))
	if err = dockerRunCmd.Run(); err != nil {
		logx.Errorf("Docker run failed: %v", err)
		os.Exit(1)
	}
	logx.Infof("Docker run completed")
}
