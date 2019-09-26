package versioncheck

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/pkg/errors"
	"github.com/tcnksm/go-latest"
)

// CheckFoGoFlutterUpdate check the last 'go-flutter' timestamp we have cached
// for the current project. If the last update comes back to more than X days,
// fetch the last Github release semver. If the Github semver is more recent
// than the current one, display the update notice.
func CheckFoGoFlutterUpdate(goDirectoryPath string, currentTag string) {
	cachedGoFlutterCheckPath := filepath.Join(goDirectoryPath, ".last_goflutter_check")
	cachedGoFlutterCheckBytes, err := ioutil.ReadFile(cachedGoFlutterCheckPath)
	if err != nil && !os.IsNotExist(err) {
		log.Warnf("Failed to read the go-flutter last update check: %v", err)
		return
	}

	cachedGoFlutterCheck := string(cachedGoFlutterCheckBytes)
	cachedGoFlutterCheck = strings.TrimSuffix(cachedGoFlutterCheck, "\n")

	now := time.Now()
	nowString := strconv.FormatInt(now.Unix(), 10)

	if cachedGoFlutterCheck == "" {
		err = ioutil.WriteFile(cachedGoFlutterCheckPath, []byte(nowString), 0664)
		if err != nil {
			log.Warnf("Failed to write the update timestamp: %v", err)
		}

		// If needed, update the hover's .gitignore file with a new entry.
		hoverGitignore := filepath.Join(goDirectoryPath, ".gitignore")
		addLineToFile(hoverGitignore, ".last_goflutter_check")

		return
	}

	i, err := strconv.ParseInt(cachedGoFlutterCheck, 10, 64)
	if err != nil {
		log.Warnf("Failed to parse the last update of go-flutter: %v", err)
		return
	}
	lastUpdateTimeStamp := time.Unix(i, 0)

	checkRate := 1.0

	newCheck := now.Sub(lastUpdateTimeStamp).Hours() > checkRate ||
		(now.Sub(lastUpdateTimeStamp).Minutes() < 1.0 && // keep the notice for X Minutes
			now.Sub(lastUpdateTimeStamp).Minutes() > 0.0)

	checkUpdateOptOut := os.Getenv("HOVER_IGNORE_CHECK_NEW_RELEASE")
	if newCheck && checkUpdateOptOut != "true" {
		log.Printf("Checking available release on Github")

		// fecth the last githubTag
		githubTag := &latest.GithubTag{
			Owner:             "go-flutter-desktop",
			Repository:        "go-flutter",
			FixVersionStrFunc: latest.DeleteFrontV(),
		}

		res, err := latest.Check(githubTag, currentTag)
		if err != nil {
			log.Warnf("Failed to check the latest release of 'go-flutter': %v", err)

			// update the timestamp
			// don't spam people who don't have access to internet
			now := time.Now().Add(time.Duration(checkRate) * time.Hour)
			nowString := strconv.FormatInt(now.Unix(), 10)

			err = ioutil.WriteFile(cachedGoFlutterCheckPath, []byte(nowString), 0664)
			if err != nil {
				log.Warnf("Failed to write the update timestamp to file: %v", err)
			}

			return
		}
		if res.Outdated {
			log.Infof("The core library 'go-flutter' has an update available. (%s -> %s)", currentTag, res.Current)
			log.Infof("              To update 'go-flutter' in this project run: $ hover upgrade")
		}

		if now.Sub(lastUpdateTimeStamp).Hours() > checkRate {
			// update the timestamp
			err = ioutil.WriteFile(cachedGoFlutterCheckPath, []byte(nowString), 0664)
			if err != nil {
				log.Warnf("Failed to write the update timestamp to file: %v", err)
			}
		}
	}

}

// CurrentGoFlutterTag retrieve the semver of go-flutter in 'go.mod'
func CurrentGoFlutterTag(goDirectoryPath string) (currentTag string, err error) {
	goModPath := filepath.Join(goDirectoryPath, "go.mod")
	goModBytes, err := ioutil.ReadFile(goModPath)
	if err != nil && !os.IsNotExist(err) {
		err = errors.Wrap(err, "Failed to read the 'go.mod' file: %v")
		return
	}

	re := regexp.MustCompile(`\sgithub.com/go-flutter-desktop/go-flutter\s(\S*)`)

	match := re.FindStringSubmatch(string(goModBytes))
	if len(match) < 2 {
		err = errors.New("Failed to parse the 'go-flutter' version in go.mod")
		return
	}
	currentTag = match[1]
	return
}

// addLineToFile appends a newLine to a file if the line isn't
// already present.
func addLineToFile(filePath, newLine string) {
	f, err := os.OpenFile(filePath,
		os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Errorf("Failed to open file %s: %v", filePath, err)
		os.Exit(1)
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		log.Errorf("Failed to read file %s: %v", filePath, err)
		os.Exit(1)
	}
	words := make(map[string]struct{})
	for _, w := range strings.Fields(strings.ToLower(string(content))) {
		words[w] = struct{}{}
	}
	_, ok := words[newLine]
	if ok {
		return
	}
	if _, err := f.WriteString(newLine + "\n"); err != nil {
		log.Errorf("Failed to append '%s' to the file (%s): %v", newLine, filePath, err)
		os.Exit(1)
	}
}
