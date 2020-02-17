package versioncheck

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tcnksm/go-latest"
	"golang.org/x/mod/modfile"

	"github.com/go-flutter-desktop/hover/internal/fileutils"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/modx"
)

func hasUpdate(timestampDir, currentVersion, repo string) (bool, string) {
	cachedCheckPath := filepath.Join(timestampDir, fmt.Sprintf(".last_%s_check", repo))
	cachedCheckBytes, err := ioutil.ReadFile(cachedCheckPath)
	if err != nil && !os.IsNotExist(err) {
		log.Warnf("Failed to read the %s last update check: %v", repo, err)
		return false, ""
	}

	cachedCheck := string(cachedCheckBytes)
	cachedCheck = strings.TrimSuffix(cachedCheck, "\n")

	now := time.Now()
	nowString := strconv.FormatInt(now.Unix(), 10)

	if cachedCheck == "" {
		err = ioutil.WriteFile(cachedCheckPath, []byte(nowString), 0664)
		if err != nil {
			log.Warnf("Failed to write the update timestamp: %v", err)
		}

		return false, ""
	}

	i, err := strconv.ParseInt(cachedCheck, 10, 64)
	if err != nil {
		log.Warnf("Failed to parse the last update of %s: %v", repo, err)
		return false, ""
	}
	lastUpdateTimeStamp := time.Unix(i, 0)

	checkRate := 1.0

	newCheck := now.Sub(lastUpdateTimeStamp).Hours() > checkRate ||
		(now.Sub(lastUpdateTimeStamp).Minutes() < 1.0 && // keep the notice for X Minutes
			now.Sub(lastUpdateTimeStamp).Minutes() > 0.0)

	checkUpdateOptOut := os.Getenv("HOVER_IGNORE_CHECK_NEW_RELEASE")
	if newCheck && checkUpdateOptOut != "true" {
		log.Printf("Checking available release on Github")

		// fetch the last githubTag
		githubTag := &latest.GithubTag{
			Owner:             "go-flutter-desktop",
			Repository:        repo,
			FixVersionStrFunc: latest.DeleteFrontV(),
		}

		res, err := latest.Check(githubTag, currentVersion)
		if err != nil {
			log.Warnf("Failed to check the latest release of '%s': %v", repo, err)

			// update the timestamp
			// don't spam people who don't have access to internet
			now := time.Now().Add(time.Duration(checkRate) * time.Hour)
			nowString := strconv.FormatInt(now.Unix(), 10)

			err = ioutil.WriteFile(cachedCheckPath, []byte(nowString), 0664)
			if err != nil {
				log.Warnf("Failed to write the update timestamp to file: %v", err)
			}

			return false, ""
		}

		if now.Sub(lastUpdateTimeStamp).Hours() > checkRate {
			// update the timestamp
			err = ioutil.WriteFile(cachedCheckPath, []byte(nowString), 0664)
			if err != nil {
				log.Warnf("Failed to write the update timestamp to file: %v", err)
			}
		}
		return res.Outdated, res.Current
	}
	return false, ""
}

// CheckForHoverUpdate check the last 'hover' timestamp we have cached.
// If the last update comes back to more than X days,
// fetch the last Github release semver. If the Github semver is more recent
// than the current one, display the update notice.
func CheckForHoverUpdate(currentVersion string) {
	// Don't check for updates if installed from local
	if currentVersion != "(devel)" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Errorf("Failed to get cache directory: %v", err)
			os.Exit(1)
		}
		update, newVersion := hasUpdate(filepath.Join(cacheDir, "hover"), currentVersion, "hover")
		if update {
			log.Infof("'hover' has an update available. (%s -> %s)", currentVersion, newVersion)
			log.Infof("              To update 'hover' go to `https://github.com/go-flutter-desktop/hover#install` and follow the install steps")
		}
	}
}

// CheckForGoFlutterUpdate check the last 'go-flutter' timestamp we have cached
// for the current project. If the last update comes back to more than X days,
// fetch the last Github release semver. If the Github semver is more recent
// than the current one, display the update notice.
func CheckForGoFlutterUpdate(goDirectoryPath string, currentTag string) {
	hoverGitignore := filepath.Join(goDirectoryPath, ".gitignore")
	fileutils.AddLineToFile(hoverGitignore, ".last_go-flutter_check")
	update, newVersion := hasUpdate(goDirectoryPath, currentTag, "go-flutter")
	if update {
		log.Infof("The core library 'go-flutter' has an update available. (%s -> %s)", currentTag, newVersion)
		log.Infof("              To update 'go-flutter' in this project run: `%s`", log.Au().Magenta("hover bumpversion"))
	}
}

// CurrentGoFlutterTag retrieve the semver of go-flutter in 'go.mod'
func CurrentGoFlutterTag(goDirectoryPath string) (currentTag string, err error) {
	const expected = "github.com/go-flutter-desktop/go-flutter"
	var m *modfile.File

	if m, err = modx.Open(goDirectoryPath); err != nil {
		return "", err
	}

	// check replacements first.
	for _, pkg := range m.Replace {
		if pkg.New.Path == expected {
			return pkg.New.Version, nil
		}
	}

	for _, pkg := range m.Require {
		if pkg.Mod.Path == expected {
			return pkg.Mod.Version, nil
		}
	}

	return "", errors.New("failed to parse the 'go-flutter' version in go.mod")
}
