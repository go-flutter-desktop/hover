package cmd

import (
	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/cmd/packaging"
)

func init() {
	initPackagingCmd.AddCommand(initLinuxSnapCmd)
	initPackagingCmd.AddCommand(initLinuxDebCmd)
	initPackagingCmd.AddCommand(initLinuxAppImageCmd)
	initPackagingCmd.AddCommand(initWindowsMsiCmd)
	initPackagingCmd.AddCommand(initDarwinBundleCmd)
	initPackagingCmd.AddCommand(initDarwinPkgCmd)
	initPackagingCmd.AddCommand(initDarwinDmgCmd)
	rootCmd.AddCommand(initPackagingCmd)
}

var initPackagingCmd = &cobra.Command{
	Use:   "init-packaging",
	Short: "Create configuration files for a packaging format",
}

var initLinuxSnapCmd = &cobra.Command{
	Use:   "linux-snap",
	Short: "Create configuration files for snap packaging",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.AssertDockerInstalled()

		packaging.InitLinuxSnap()
	},
}

var initLinuxDebCmd = &cobra.Command{
	Use:   "linux-deb",
	Short: "Create configuration files for deb packaging",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.AssertDockerInstalled()

		packaging.InitLinuxDeb()
	},
}

var initLinuxAppImageCmd = &cobra.Command{
	Use:   "linux-appimage",
	Short: "Create configuration files for AppImage packaging",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.AssertDockerInstalled()

		packaging.InitLinuxAppImage()
	},
}

var initWindowsMsiCmd = &cobra.Command{
	Use:   "windows-msi",
	Short: "Create configuration files for msi packaging",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.AssertDockerInstalled()

		packaging.InitWindowsMsi()
	},
}

var initDarwinBundleCmd = &cobra.Command{
	Use:   "darwin-bundle",
	Short: "Create configuration files for OSX bundle packaging",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.AssertDockerInstalled()

		packaging.InitDarwinBundle()
	},
}

var initDarwinPkgCmd = &cobra.Command{
	Use:   "darwin-pkg",
	Short: "Create configuration files for OSX pkg installer packaging",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.AssertDockerInstalled()

		packaging.InitDarwinPkg()
	},
}

var initDarwinDmgCmd = &cobra.Command{
	Use:   "darwin-dmg",
	Short: "Create configuration files for OSX dmg packaging",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.AssertDockerInstalled()

		packaging.InitDarwinDmg()
	},
}
