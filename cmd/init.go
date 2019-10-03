package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
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
				log.Errorf("A file or directory named '%s' already exists. Cannot continue init.", buildPath)
				os.Exit(1)
			}
			log.Errorf("Failed to create '%s' directory: %v", buildPath, err)
			os.Exit(1)
		}

		desktopCmdPath := filepath.Join(buildPath, "cmd")
		err = os.Mkdir(desktopCmdPath, 0775)
		if err != nil {
			log.Errorf("Failed to create '%s': %v", desktopCmdPath, err)
			os.Exit(1)
		}

		desktopAssetsPath := filepath.Join(buildPath, "assets")
		err = os.Mkdir(desktopAssetsPath, 0775)
		if err != nil {
			log.Errorf("Failed to create '%s': %v", desktopAssetsPath, err)
			os.Exit(1)
		}

		fileutils.CopyAsset("app/main.go", filepath.Join(desktopCmdPath, "main.go"), assetsBox)
		fileutils.CopyAsset("app/options.go", filepath.Join(desktopCmdPath, "options.go"), assetsBox)
		fileutils.CopyAsset("app/icon.png", filepath.Join(desktopAssetsPath, "icon.png"), assetsBox)
		fileutils.CopyAsset("app/gitignore", filepath.Join(buildPath, ".gitignore"), assetsBox)

		initializeGoModule(projectPath)
		log.Printf("Available plugin for this project:")
		pluginListCmd.Run(cmd, []string{})
	},
}
