package packaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

// DarwinPkgTask packaging for darwin as pkg
var DarwinPkgTask = &packagingTask{
	packagingFormatName: "darwin-pkg",
	dependsOn: map[*packagingTask]string{
		DarwinBundleTask: "flat/root/Applications",
	},
	templateFiles: map[string]string{
		"darwin-pkg/PackageInfo.tmpl":  "flat/base.pkg/PackageInfo.tmpl",
		"darwin-pkg/Distribution.tmpl": "flat/Distribution.tmpl",
	},
	packagingFunction: func(tmpPath, applicationName, packageName, executableName, version, release string) (string, error) {
		outputFileName := fmt.Sprintf("%s %s.pkg", applicationName, version)

		payload, err := os.OpenFile(filepath.Join(tmpPath, "flat", "base.pkg", "Payload"), os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return "", err
		}

		appBundleOriginalPath := filepath.Join(tmpPath, "flat", "root", "Applications", fmt.Sprintf("%s %s.app", applicationName, version))
		appBundleFinalPath := filepath.Join(tmpPath, "flat", "root", "Applications", fmt.Sprintf("%s.app", applicationName))
		err = os.Rename(appBundleOriginalPath, appBundleFinalPath)
		if err != nil {
			return "", err
		}

		cmdFind := exec.Command("find", ".")
		cmdFind.Dir = filepath.Join(tmpPath, "flat", "root")
		cmdCpio := exec.Command("cpio", "-o", "--format", "odc", "--owner", "0:80")
		cmdCpio.Dir = filepath.Join(tmpPath, "flat", "root")
		cmdGzip := exec.Command("gzip", "-c")

		// Pipes like this: find | cpio | gzip > Payload
		cmdCpio.Stdin, err = cmdFind.StderrPipe()
		if err != nil {
			return "", err
		}
		cmdGzip.Stdin, err = cmdCpio.StderrPipe()
		if err != nil {
			return "", err
		}
		cmdGzip.Stdout = payload

		err = cmdGzip.Start()
		if err != nil {
			return "", err
		}
		err = cmdCpio.Start()
		if err != nil {
			return "", err
		}
		err = cmdFind.Run()
		if err != nil {
			return "", err
		}
		err = cmdCpio.Wait()
		if err != nil {
			return "", err
		}
		err = cmdGzip.Wait()
		if err != nil {
			return "", err
		}
		err = payload.Close()
		if err != nil {
			return "", err
		}

		var cmdMkbom *exec.Cmd
		switch os := runtime.GOOS; os {
		case "darwin":
			cmdMkbom = exec.Command("mkbom", filepath.Join("flat", "root"), filepath.Join("flat", "base.pkg", "Payload"))
		case "linux":
			cmdMkbom = exec.Command("mkbom", "-u", "0", "-g", "80", filepath.Join("flat", "root"), filepath.Join("flat", "base.pkg", "Payload"))
		}
		cmdMkbom.Dir = tmpPath
		cmdMkbom.Stdout = os.Stdout
		cmdMkbom.Stderr = os.Stderr
		err = cmdMkbom.Run()
		if err != nil {
			return "", nil
		}

		var files []string
		err = filepath.Walk(filepath.Join(tmpPath, "flat"), func(path string, info os.FileInfo, err error) error {
			relativePath, err := filepath.Rel(filepath.Join(tmpPath, "flat"), path)
			if err != nil {
				return err
			}
			files = append(files, relativePath)
			return nil
		})
		if err != nil {
			return "", errors.Wrap(err, "failed to iterate over ")
		}

		cmdXar := exec.Command("xar", append([]string{"--compression", "none", "-cf", filepath.Join("..", outputFileName)}, files...)...)
		cmdXar.Dir = filepath.Join(tmpPath, "flat")
		cmdXar.Stdout = os.Stdout
		cmdXar.Stderr = os.Stderr
		err = cmdXar.Run()
		if err != nil {
			return "", errors.Wrap(err, "failed to run xar")
		}
		return outputFileName, nil
	},
	requiredTools: map[string]map[string]string{
		"linux": {
			"find":  "Install find from your package manager",
			"cpio":  "Install cpio from your package manager",
			"gzip":  "Install gzip from your package manager",
			"mkbom": "Install bomutils from your package manager or from https://github.com/hogliux/bomutils",
			"xar":   "Install xar from your package manager or from https://github.com/mackyle/xar",
		},
		"darwin": {
			"find":  "Install find from your package manager",
			"cpio":  "Install cpio from your package manager",
			"gzip":  "Install gzip from your package manager",
			"mkbom": "Install bomutils from your package manager or from https://github.com/hogliux/bomutils",
			"xar":   "Install xar from your package manager or from https://github.com/mackyle/xar",
		},
	},
}
