package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/config"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
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

		projectName := pubspec.GetPubSpec().Name

		var projectPath string
		if len(args) == 0 || args[0] == "." {
			projectPath = projectName
		} else {
			projectPath = args[0]
		}

		err := os.Mkdir(build.BuildPath, 0775)
		if err != nil {
			if os.IsExist(err) {
				log.Errorf("A file or directory named '%s' already exists. Cannot continue init.", build.BuildPath)
				os.Exit(1)
			}
			log.Errorf("Failed to create '%s' directory: %v", build.BuildPath, err)
			os.Exit(1)
		}

		desktopCmdPath := filepath.Join(build.BuildPath, "cmd")
		err = os.Mkdir(desktopCmdPath, 0775)
		if err != nil {
			log.Errorf("Failed to create '%s': %v", desktopCmdPath, err)
			os.Exit(1)
		}

		desktopAssetsPath := filepath.Join(build.BuildPath, "assets")
		err = os.Mkdir(desktopAssetsPath, 0775)
		if err != nil {
			log.Errorf("Failed to create '%s': %v", desktopAssetsPath, err)
			os.Exit(1)
		}

		emptyConfig := config.Config{}

		fileutils.CopyAsset("app/main.go", filepath.Join(desktopCmdPath, "main.go"), fileutils.AssetsBox())
		fileutils.CopyAsset("app/options.go", filepath.Join(desktopCmdPath, "options.go"), fileutils.AssetsBox())
		fileutils.CopyAsset("app/icon.png", filepath.Join(desktopAssetsPath, "icon.png"), fileutils.AssetsBox())
		fileutils.CopyAsset("app/gitignore", filepath.Join(build.BuildPath, ".gitignore"), fileutils.AssetsBox())
		fileutils.ExecuteTemplateFromAssetsBox("app/hover.yaml.tmpl", filepath.Join(build.BuildPath, "hover.yaml"), fileutils.AssetsBox(), map[string]string{
			"applicationName": emptyConfig.GetApplicationName(projectName),
			"executableName":  emptyConfig.GetExecutableName(projectName),
			"packageName":     emptyConfig.GetPackageName(projectName),
		})

		initializeGoModule(projectPath)
		log.Printf("Available plugin for this project:")
		pluginListCmd.Run(cmd, []string{})
	},
}
