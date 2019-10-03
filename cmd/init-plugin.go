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
	rootCmd.AddCommand(createPluginCmd)
}

var createPluginCmd = &cobra.Command{
	Use:   "init-plugin",
	Short: "Initialize a go-flutter plugin in a existing flutter platform plugin",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires one argument, the VCS repository path. e.g.: github.com/my-organization/" + getPubSpec().Name + "\n" +
				"This path will be used by Golang to fetch the plugin, make sure it correspond to the code repository of the plugin!")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterPluginProject()

		vcsPath := args[0]

		err := os.Mkdir(buildPath, 0775)
		if err != nil {
			if os.IsExist(err) {
				log.Errorf("A file or directory named `" + buildPath + "` already exists. Cannot continue init-plugin.")
				os.Exit(1)
			}
			log.Errorf("Failed to create '%s' directory: %v", buildPath, err)
			os.Exit(1)
		}

		templateData := map[string]string{
			"pluginName": getPubSpec().Name,
			"structName": toCamelCase(getPubSpec().Name + "Plugin"),
			"urlVSCRepo": vcsPath,
		}

		fileutils.CopyTemplate("plugin/plugin.go.tmpl", filepath.Join(buildPath, "plugin.go"), assetsBox, templateData)
		fileutils.CopyTemplate("plugin/README.md.tmpl", filepath.Join(buildPath, "README.md"), assetsBox, templateData)
		fileutils.CopyTemplate("plugin/import.go.tmpl.tmpl", filepath.Join(buildPath, "import.go.tmpl"), assetsBox, templateData)

		initializeGoModule(vcsPath)
	},
}
