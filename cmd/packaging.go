package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/go-flutter-desktop/hover/cmd/packaging"
	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/log"
)

var (
	packagingTargetsString string
)

func init() {
	initPackagingCmd.Flags().StringVar(&packagingTargetsString, "packaging-targets", "", "The packaging targets to init packaging for")
	initPackagingCmd.MarkFlagRequired("packaging-targets")
	initPackagingCmd.AddCommand(listPackagingCmd)
	initPackagingCmd.AddCommand(initLinuxSnapCmd)
	initPackagingCmd.AddCommand(initLinuxDebCmd)
	initPackagingCmd.AddCommand(initLinuxAppImageCmd)
	initPackagingCmd.AddCommand(initLinuxRpmCmd)
	initPackagingCmd.AddCommand(initWindowsMsiCmd)
	initPackagingCmd.AddCommand(initDarwinBundleCmd)
	initPackagingCmd.AddCommand(initDarwinPkgCmd)
	initPackagingCmd.AddCommand(initDarwinDmgCmd)
	rootCmd.AddCommand(initPackagingCmd)
}

var listPackagingCmd = &cobra.Command{
	Use:   "list",
	Short: "List all packaging targets",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("All packaging targets:")
		for _, buildTarget := range build.PackagingTargets {
			line := fmt.Sprintf("    %s-%s", buildTarget.Platform, buildTarget.PackagingFormat)
			_, err := os.Stat(packaging.PackagingFormatPath(buildTarget))
			initialized := err == nil
			if build.DockerBin == "" || !initialized {
				line += " ("
				if build.DockerBin == "" {
					line += "docker needs to be installed"
				}
				if build.DockerBin == "" && !initialized {
					line += "; "
				}
				if !initialized {
					line += "needs to be initialized first"
				}
				line += ")"
			}
			log.Infof(line)
		}
	},
}

var initLinuxSnapCmd = &cobra.Command{
	Use:   "linux-snap",
	Short: "Create configuration files for snap packaging",
	Run: func(cmd *cobra.Command, args []string) {
		initPackagingFromTargets([]build.Target{{
			Platform:        build.TargetPlatforms.Linux,
			PackagingFormat: build.TargetPackagingFormats.Snap,
		}})
	},
}

var initLinuxDebCmd = &cobra.Command{
	Use:   "linux-deb",
	Short: "Create configuration files for deb packaging",
	Run: func(cmd *cobra.Command, args []string) {
		initPackagingFromTargets([]build.Target{{
			Platform:        build.TargetPlatforms.Linux,
			PackagingFormat: build.TargetPackagingFormats.Deb,
		}})
	},
}

var initLinuxAppImageCmd = &cobra.Command{
	Use:   "linux-appimage",
	Short: "Create configuration files for AppImage packaging",
	Run: func(cmd *cobra.Command, args []string) {
		initPackagingFromTargets([]build.Target{{
			Platform:        build.TargetPlatforms.Linux,
			PackagingFormat: build.TargetPackagingFormats.AppImage,
		}})
	},
}
var initLinuxRpmCmd = &cobra.Command{
	Use:   "linux-rpm",
	Short: "Create configuration files for rpm packaging",
	Run: func(cmd *cobra.Command, args []string) {
		initPackagingFromTargets([]build.Target{{
			Platform:        build.TargetPlatforms.Linux,
			PackagingFormat: build.TargetPackagingFormats.Rpm,
		}})
	},
}

var initDarwinBundleCmd = &cobra.Command{
	Use:   "darwin-bundle",
	Short: "Create configuration files for OSX bundle packaging",
	Run: func(cmd *cobra.Command, args []string) {
		initPackagingFromTargets([]build.Target{{
			Platform:        build.TargetPlatforms.Darwin,
			PackagingFormat: build.TargetPackagingFormats.Bundle,
		}})
	},
}

var initDarwinPkgCmd = &cobra.Command{
	Use:   "darwin-pkg",
	Short: "Create configuration files for OSX pkg installer packaging",
	Run: func(cmd *cobra.Command, args []string) {
		initPackagingFromTargets([]build.Target{{
			Platform:        build.TargetPlatforms.Darwin,
			PackagingFormat: build.TargetPackagingFormats.Pkg,
		}})
	},
}

var initDarwinDmgCmd = &cobra.Command{
	Use:   "darwin-dmg",
	Short: "Create configuration files for OSX dmg packaging",
	Run: func(cmd *cobra.Command, args []string) {
		initPackagingFromTargets([]build.Target{{
			Platform:        build.TargetPlatforms.Darwin,
			PackagingFormat: build.TargetPackagingFormats.Dmg,
		}})
	},
}

var initWindowsMsiCmd = &cobra.Command{
	Use:   "windows-msi",
	Short: "Create configuration files for msi packaging",
	Run: func(cmd *cobra.Command, args []string) {
		initPackagingFromTargets([]build.Target{{
			Platform:        build.TargetPlatforms.Windows,
			PackagingFormat: build.TargetPackagingFormats.Msi,
		}})
	},
}

var initPackagingCmd = &cobra.Command{
	Use:   "init-packaging",
	Short: "Create configuration files for a packaging format",
	Run: func(cmd *cobra.Command, args []string) {
		buildTargets, err := build.ParseBuildTargets(packagingTargetsString, true)
		if err != nil {
			log.Errorf("Failed to parse packaging targets: %v", err)
			os.Exit(1)
		}
		initPackagingFromTargets(buildTargets)
	},
}

func initPackagingFromTargets(buildTargets []build.Target) {
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
}
