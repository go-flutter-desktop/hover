package enginecache

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	latest "github.com/tcnksm/go-latest"
)

func flutterRequiredEngineVersion() string {
	out, err := exec.Command("flutter", "--version").Output()
	if err != nil {
		fmt.Printf("hover: Failed to run `flutter --version`: %v\n", err)
		os.Exit(1)
	}

	regexpEngineVersion := regexp.MustCompile(`Engine • revision (\w{10})`)
	versionMatch := regexpEngineVersion.FindStringSubmatch(string(out))
	if len(versionMatch) != 2 {
		fmt.Printf("hover: Failed to obtain engine version")
		os.Exit(1)
	}

	return versionMatch[1]
}

// CheckFoGoFlutterUpdate check the last 'go-flutter' timestamp we have cached
// for the current project. If the last update comes back to more than X days,
// fetch the last Github release semver. If the Github semver is more recent
// than the current one, display the update notice.
func CheckFoGoFlutterUpdate(wd string, currentTag string) {
	cachedGoFlutterCheckPath := filepath.Join(wd, ".last_goflutter_check")
	cachedGoFlutterCheckBytes, err := ioutil.ReadFile(cachedGoFlutterCheckPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("hover: Failed to read the go-flutter last update check: %v\n", err)
		return
	}

	cachedGoFlutterCheck := string(cachedGoFlutterCheckBytes)
	cachedGoFlutterCheck = strings.TrimSuffix(cachedGoFlutterCheck, "\n")

	now := time.Now()
	nowString := strconv.FormatInt(now.Unix(), 10)

	if cachedGoFlutterCheck == "" {
		err = ioutil.WriteFile(cachedGoFlutterCheckPath, []byte(nowString), 0664)
		if err != nil {
			fmt.Printf("hover: Failed to write the update timestamp to file: %v\n", err)
		}
		return
	}

	i, err := strconv.ParseInt(cachedGoFlutterCheck, 10, 64)
	if err != nil {
		fmt.Printf("hover: Failed to parse the last update of go-flutter: %v\n", err)
		return
	}
	lastUpdateTimeStamp := time.Unix(i, 0)

	newCheck := now.Sub(lastUpdateTimeStamp).Hours() > 48.0 ||
		(now.Sub(lastUpdateTimeStamp).Minutes() < 1.0 && // keep the notice for X Minutes
			now.Sub(lastUpdateTimeStamp).Minutes() > 0.0)

	if newCheck {
		fmt.Printf("hover: Checking available release on Github\n")

		// fecth the last githubTag
		githubTag := &latest.GithubTag{
			Owner:             "go-flutter-desktop",
			Repository:        "go-flutter",
			FixVersionStrFunc: latest.DeleteFrontV(),
		}

		res, err := latest.Check(githubTag, currentTag)
		if err != nil {
			fmt.Printf("hover: Failed to check the latest release of 'go-flutter': %v\n", err)

			// update the timestamp
			// don't spam people who don't have access to internet
			now := time.Now().Add(48 * time.Hour)
			nowString := strconv.FormatInt(now.Unix(), 10)

			err = ioutil.WriteFile(cachedGoFlutterCheckPath, []byte(nowString), 0664)
			if err != nil {
				fmt.Printf("hover: Failed to write the update timestamp to file: %v\n", err)
			}

			return
		}
		if res.Outdated {
			fmt.Printf("hover: The core library 'go-flutter' has an update available. (%s -> %s)\n", currentTag, res.Current)
			fmt.Printf("              To update 'go-flutter' run: $ hover upgrade\n")
		}

		if now.Sub(lastUpdateTimeStamp).Hours() > 48.0 {
			// update the timestamp
			err = ioutil.WriteFile(cachedGoFlutterCheckPath, []byte(nowString), 0664)
			if err != nil {
				fmt.Printf("hover: Failed to write the update timestamp to file: %v\n", err)
			}
		}
	}

}

// CurrentGoFlutterTag retrieve the semver of go-flutter in 'go.mod'
func CurrentGoFlutterTag(wd string) (currentTag string, err error) {
	goModPath := filepath.Join(wd, "go.mod")
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
