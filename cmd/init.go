package cmd

import (
	"errors"
	"github.com/go-flutter-desktop/hover/internal/log"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [project]",
	Short: "Initialize a flutter project to use go-flutter",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("allows only one argument, the project path. e.g.: github.com/my-organization/my-app")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterProject()

		var projectPath string
		if len(args) == 0 || args[0] == "." {
			projectPath = getPubSpec().Name
		} else {
			projectPath = args[0]
		}

		err := os.Mkdir(buildPath, 0775)
		if err != nil {
			if os.IsExist(err) {
				log.Fatal("A file or directory named `" + buildPath + "` already exists. Cannot continue init.")
				os.Exit(1)
			}
		}

		desktopCmdPath := filepath.Join(buildPath, "cmd")
		err = os.Mkdir(desktopCmdPath, 0775)
		if err != nil {
			log.Fatal("Failed to create `%s`: %v", desktopCmdPath, err)
			os.Exit(1)
		}

		desktopAssetsPath := filepath.Join(buildPath, "assets")
		err = os.Mkdir(desktopAssetsPath, 0775)
		if err != nil {
			log.Fatal("Failed to create `%s`: %v", desktopAssetsPath, err)
			os.Exit(1)
		}

		copyAsset("app/main.go", filepath.Join(desktopCmdPath, "main.go"))
		copyAsset("app/options.go", filepath.Join(desktopCmdPath, "options.go"))
		copyAsset("app/icon.png", filepath.Join(desktopAssetsPath, "icon.png"))
		copyAsset("app/gitignore", filepath.Join(buildPath, ".gitignore"))

		wd, err := os.Getwd()
		if err != nil {
			log.Fatal("Failed to get working dir: %v", err)
			os.Exit(1)
		}

		cmdGoModInit := exec.Command(goBin, "mod", "init", projectPath+"/"+buildPath)
		cmdGoModInit.Dir = filepath.Join(wd, buildPath)
		cmdGoModInit.Env = append(os.Environ(),
			"GO111MODULE=on",
		)
		cmdGoModInit.Stderr = os.Stderr
		cmdGoModInit.Stdout = os.Stdout
		err = cmdGoModInit.Run()
		if err != nil {
			log.Fatal("Go mod init failed: %v", err)
			os.Exit(1)
		}

		cmdGoModTidy := exec.Command(goBin, "mod", "tidy")
		cmdGoModTidy.Dir = filepath.Join(wd, buildPath)
		log.Print(cmdGoModTidy.Dir)
		cmdGoModTidy.Env = append(os.Environ(),
			"GO111MODULE=on",
		)
		cmdGoModTidy.Stderr = os.Stderr
		cmdGoModTidy.Stdout = os.Stdout
		err = cmdGoModTidy.Run()
		if err != nil {
			log.Fatal("Go mod tidy failed: %v", err)
			os.Exit(1)
		}
	},
}

func copyAsset(boxed, to string) {
	file, err := os.Create(to)
	if err != nil {
		log.Fatal("Failed to create %s: %v", to, err)
		os.Exit(1)
	}
	defer file.Close()
	boxedFile, err := assetsBox.Open(boxed)
	if err != nil {
		log.Fatal("Failed to find boxed file %s: %v", boxed, err)
		os.Exit(1)
	}
	defer boxedFile.Close()
	_, err = io.Copy(file, boxedFile)
	if err != nil {
		log.Fatal("Failed to write file %s: %v", to, err)
		os.Exit(1)
	}
}
