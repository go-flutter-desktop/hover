package packaging

import (
	"fmt"
	"os"
	"os/exec"
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
		appFileOriginalPath := fmt.Sprintf("dmgdir/%s %s.app", applicationName, version)
		appFileFinalPath := fmt.Sprintf("dmgdir/%s.app", applicationName)
		cmdRenameApp := exec.Command("mv", appFileOriginalPath, appFileFinalPath)
		cmdRenameApp.Dir = tmpPath
		cmdRenameApp.Stdout = os.Stdout
		cmdRenameApp.Stderr = os.Stderr
		err = cmdRenameApp.Run()
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
