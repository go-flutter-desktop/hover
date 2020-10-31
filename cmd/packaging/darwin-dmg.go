package packaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

		var cmdCreateBundle *exec.Cmd
		switch os := runtime.GOOS; os {
		case "darwin":
			cmdCreateBundle = exec.Command("hdiutil", "create", "-volname", packageName, "-srcfolder", "dmgdir", "-ov", "-format", "UDBZ", outputFileName)
		case "linux":
			cmdCreateBundle = exec.Command("mkisofs", "-V", packageName, "-D", "-R", "-apple", "-no-pad", "-o", outputFileName, "dmgdir")
		}
		cmdCreateBundle.Dir = tmpPath
		cmdCreateBundle.Stdout = os.Stdout
		cmdCreateBundle.Stderr = os.Stderr
		err = cmdCreateBundle.Run()
		if err != nil {
			return "", err
		}
		return outputFileName, nil
	},
	skipAssertInitialized: true,
	requiredTools: map[string]map[string]string{
		"linux": {
			"ln":      "Install ln from your package manager",
			"mkisofs": "Install mkisofs from your package manager. Some distros ship genisoimage which is a fork of mkisofs. Create a symlink for it like this: ln -s $(which genisoimage) /usr/bin/mkisofs",
		},
		"darwin": {
			"ln":      "Install ln from your package manager",
			"hdiutil": "Install hdiutil from your package manager",
		},
	},
}
