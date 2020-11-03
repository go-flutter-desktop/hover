package darwinhacks

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-flutter-desktop/hover/internal/log"
)

// DyldHack is a nasty hack to get the linking working. After fiddling a lot of hours with CGO linking
// this was the only solution I could come up with and it works. I guess something would need to be changed in the engine
// builds to make this obsolete, but this hack does it for now.
func DyldHack(path string) {
	installNameToolCommand := []string{
		"install_name_tool",
		"-change",
		"./libflutter_engine.dylib",
		"@executable_path/libflutter_engine.dylib",
		"-id",
		"@executable_path/libflutter_engine.dylib",
		RewriteDarlingPath(runtime.GOOS != "darwin", path),
	}
	if runtime.GOOS != "darwin" {
		installNameToolCommand = append([]string{"darling", "shell"}, installNameToolCommand...)
	}
	cmdInstallNameTool := exec.Command(
		installNameToolCommand[0],
		installNameToolCommand[1:]...,
	)
	cmdInstallNameTool.Stderr = os.Stderr
	output, err := cmdInstallNameTool.Output()
	if err != nil {
		log.Errorf("install_name_tool failed: %v", err)
		log.Errorf(string(output))
		os.Exit(1)
	}
}

func RewriteDarlingPath(useDarling bool, path string) string {
	if useDarling {
		return filepath.Join("/", "Volumes", "SystemRoot", path)
	}
	return path
}

func ChangePackagesFilePath(isInsert bool) {
	for _, path := range []string{".packages", filepath.Join(".dart_tool", "package_config.json")} {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Errorf("Failed to read %s file: %v", path, err)
			os.Exit(1)
		}
		lines := strings.Split(string(content), "\n")
		for i := range lines {
			if strings.Contains(lines[i], "file://") {
				parts := strings.Split(lines[i], "file://")
				if isInsert && !strings.Contains(lines[i], "/Volumes/SystemRoot") {
					lines[i] = fmt.Sprintf("%sfile:///Volumes/SystemRoot%s", parts[0], parts[1])
				} else {
					lines[i] = fmt.Sprintf("%sfile://%s", parts[0], strings.ReplaceAll(parts[1], "/Volumes/SystemRoot", ""))
				}
			}
		}
		err = ioutil.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			log.Errorf("Failed to write %s file: %v", path, err)
			os.Exit(1)
		}
	}
}
