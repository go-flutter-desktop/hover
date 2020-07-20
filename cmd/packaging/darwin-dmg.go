package packaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DarwinDmgTask packaging for darwin as dmg
var DarwinDmgTask = &packagingTask{
	packagingFormatName: "darwin-dmg",
	dependsOn: map[*packagingTask]string{
		DarwinBundleTask: "dmgdir",
	},
	packagingFunction: func(tmpPath, applicationName, packageName, executableName, version, release string) (string, error) {
		outputFileName := fmt.Sprintf("%s %s.dmg", applicationName, version)
		cmdLn := exec.Command("ln", "-sf", "/Applications", "dmgdir/Applications")
		cmdLn.Dir = tmpPath
		cmdLn.Stdout = os.Stdout
		cmdLn.Stderr = os.Stderr
		err := cmdLn.Run()
		if err != nil {
			return "", err
		}
		appBundleOriginalPath := filepath.Join(tmpPath, "dmgdir", fmt.Sprintf("%s %s.app", applicationName, version))
		appBundleFinalPath := filepath.Join(tmpPath, "dmgdir", fmt.Sprintf("%s.app", applicationName))
		err = os.Rename(appBundleOriginalPath, appBundleFinalPath)
		if err != nil {
			return "", err
		}
		cmdGenisoimage := exec.Command("genisoimage", "-V", packageName, "-D", "-R", "-apple", "-no-pad", "-o", outputFileName, "dmgdir")
		cmdGenisoimage.Dir = tmpPath
		cmdGenisoimage.Stdout = os.Stdout
		cmdGenisoimage.Stderr = os.Stderr
		err = cmdGenisoimage.Run()
		if err != nil {
			return "", err
		}
		return outputFileName, nil
	},
	skipAssertInitialized: true,
	requiredTools: map[string][]string{
		"linux": {"ln", "genisoimage"},
	},
}
