package cmd

import (
	"github.com/spf13/cobra"

	"github.com/go-flutter-desktop/hover/cmd/packaging"
)

func init() {
	initPackagingCmd.AddCommand(initLinuxSnapCmd)
	initPackagingCmd.AddCommand(initLinuxDebCmd)
	initPackagingCmd.AddCommand(initLinuxAppImageCmd)
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
		packaging.DockerInstalled()

		packaging.InitLinuxSnap()
	},
}

var initLinuxDebCmd = &cobra.Command{
	Use:   "linux-deb",
	Short: "Create configuration files for deb packaging",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.DockerInstalled()

		packaging.InitLinuxDeb()
	},
}

var initLinuxAppImageCmd = &cobra.Command{
	Use:   "linux-appimage",
	Short: "Create configuration files for AppImage packaging",
	Run: func(cmd *cobra.Command, args []string) {
		assertHoverInitialized()
		packaging.DockerInstalled()

		packaging.InitLinuxAppImage()
	},
}
