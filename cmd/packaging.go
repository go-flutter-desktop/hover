package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/cmd/packaging"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
)

func init() {
	rootCmd.AddCommand(initPackagingCmd)
}

var initPackagingCmd = &cobra.Command{
	Use:   "init-packaging",
	Short: "Create configuration files for a packaging format",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a build targets argument")
		}
		return build.AreValidBuildTargets(args[0], true)
	},
	Run: func(cmd *cobra.Command, args []string) {
		buildTargets, err := build.ParseBuildTargets(args[0], true)
		if err != nil {
			log.Errorf("Failed to parse build targets: %v", err)
			os.Exit(1)
		}
		assertHoverInitialized()
		packaging.AssertDockerInstalled()
		for _, buildTarget := range buildTargets {
			if buildTarget.PackagingFormat == build.TargetPackagingFormats.AppImage {
				packaging.InitLinuxAppImage(buildTarget)
			} else if buildTarget.PackagingFormat == build.TargetPackagingFormats.Deb {
				packaging.InitLinuxDeb(buildTarget)
			} else if buildTarget.PackagingFormat == build.TargetPackagingFormats.Snap {
				packaging.InitLinuxSnap(buildTarget)
			} else if buildTarget.PackagingFormat == build.TargetPackagingFormats.Bundle {
				packaging.InitDarwinBundle(buildTarget)
			} else if buildTarget.PackagingFormat == build.TargetPackagingFormats.Pkg {
				packaging.InitDarwinPkg(buildTarget)
			} else if buildTarget.PackagingFormat == build.TargetPackagingFormats.Msi {
				packaging.InitWindowsMsi(buildTarget)
			}
		}
	},
}
