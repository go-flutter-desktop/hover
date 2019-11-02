package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/pubspec"
)

func init() {
	rootCmd.AddCommand(createPluginCmd)
}

var createPluginCmd = &cobra.Command{
	Use:   "init-plugin",
	Short: "Initialize a go-flutter plugin in a existing flutter platform plugin",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires one argument, the VCS repository path. e.g.: github.com/my-organization/" + pubspec.GetPubSpec().Name + "\n" +
				"This path will be used by Golang to fetch the plugin, make sure it correspond to the code repository of the plugin!")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterPluginProject()

		vcsPath := args[0]

		err := os.Mkdir(build.BuildPath, 0775)
		if err != nil {
			if os.IsExist(err) {
				log.Errorf("A file or directory named `" + build.BuildPath + "` already exists. Cannot continue init-plugin.")
				os.Exit(1)
			}
			log.Errorf("Failed to create `%s` directory: %v", build.BuildPath, err)
			os.Exit(1)
		}

		templateData := map[string]string{
			"pluginName": pubspec.GetPubSpec().Name,
			"structName": toCamelCase(pubspec.GetPubSpec().Name + "Plugin"),
			"urlVSCRepo": vcsPath,
		}

		fileutils.CopyTemplate("plugin/plugin.go.tmpl", filepath.Join(build.BuildPath, "plugin.go"), fileutils.AssetsBox, templateData)
		fileutils.CopyTemplate("plugin/README.md.tmpl", filepath.Join(build.BuildPath, "README.md"), fileutils.AssetsBox, templateData)
		fileutils.CopyTemplate("plugin/import.go.tmpl.tmpl", filepath.Join(build.BuildPath, "import.go.tmpl"), fileutils.AssetsBox, templateData)

		initializeGoModule(vcsPath)
	},
}
