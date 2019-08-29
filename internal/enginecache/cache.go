package enginecache

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func createSymLink(oldname, newname string) error {
	err := os.Remove(newname)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove existing symlink")
	}

	err = os.Symlink(oldname, newname)
	if err != nil {
		return errors.Wrap(err, "failed to create symlink")
	}
	return nil
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func unzip(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// Function to prind download percent completion
func printDownloadPercent(done chan chan struct{}, path string, expectedSize int64) {
	var completedCh chan struct{}
	for {
		fi, err := os.Stat(path)
		if err != nil {
			log.Fatal(err)
		}

		size := fi.Size()

		if size == 0 {
			size = 1
		}

		var percent = float64(size) / float64(expectedSize) * 100

		// We use `\033[2K\r` to avoid carriage return, it will print above previous.
		fmt.Printf("\033[2K\r %.0f %% / 100 %%", percent)

		if completedCh != nil {
			close(completedCh)
			return
		}

		select {
		case completedCh = <-done:
		case <-time.After(time.Second / 60): // Flutter promises 60fps, right? ;)
		}
	}
}

func moveFile(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("Couldn't open src file: %s", err)
	}
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := srcFileInfo.Mode() & os.ModePerm
	destFile, err := os.OpenFile(destPath, flag, perm)
	if err != nil {
		srcFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, srcFile)
	srcFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(srcPath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}

// Function to download file with given path and url.
func downloadFile(filepath string, url string) error {
	// // Print download url in case user needs it.
	// fmt.Printf("Downloading file from\n '%s'\n to '%s'\n\n", url, filepath)

	start := time.Now()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	expectedSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return errors.Wrap(err, "failed to get Content-Length header")
	}

	doneCh := make(chan chan struct{})
	go printDownloadPercent(doneCh, filepath, int64(expectedSize))

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// close channel to indicate we're done
	doneCompletedCh := make(chan struct{})
	doneCh <- doneCompletedCh // signal that download is done
	<-doneCompletedCh         // wait for signal that printing has completed

	elapsed := time.Since(start)
	log.Printf("\033[2K\rhover: Download completed in %.2fs\n", elapsed.Seconds())
	return nil
}

// ValidateOrUpdateEngineAtPath validates the engine we have cached matches the
// flutter version, or otherwise downloads a new engine. The engine cache
// location is set by the the user.
func ValidateOrUpdateEngineAtPath(targetOS string, cachePath string) (engineCachePath string) {
	engineCachePath = filepath.Join(cachePath, "hover", "engine", targetOS)

	if strings.Contains(engineCachePath, " ") {
		fmt.Printf("hover: Cannot save the engine to '%s', engine cache is not compatible with path containing spaces.\n", cachePath)
		fmt.Printf("       Please run hover with a another engine cache path. Example:\n")
		fmt.Printf("              hover run --cache-path \"C:\\cache\"\n\n")
		fmt.Printf("       The --cache-path flag will have to be provided to every build and run command.")
		os.Exit(1)
	}

	cachedEngineVersionPath := filepath.Join(engineCachePath, "version")
	cachedEngineVersionBytes, err := ioutil.ReadFile(cachedEngineVersionPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("hover: Failed to read cached engine version: %v\n", err)
		os.Exit(1)
	}
	cachedEngineVersion := string(cachedEngineVersionBytes)
	requiredEngineVersion := flutterRequiredEngineVersion()

	if cachedEngineVersion != "" {
		if cachedEngineVersion == requiredEngineVersion {
			fmt.Println("hover: Using engine from cache")
			return
		}

		// Engine is outdated, we remove the old engine and continue to download
		// the new engine.
		err = os.RemoveAll(engineCachePath)
		if err != nil {
			fmt.Printf("hover: Failed to remove outdated engine: %v\n", err)
			os.Exit(1)
		}
	}

	err = os.MkdirAll(engineCachePath, 0775)
	if err != nil {
		fmt.Printf("hover: Failed to create engine cache directory: %v\n", err)
		os.Exit(1)
	}

	targetedDomain := "https://storage.googleapis.com"
	envURLFlutter := os.Getenv("FLUTTER_STORAGE_BASE_URL")
	if envURLFlutter != "" {
		targetedDomain = envURLFlutter
	}

	// Retrieve the full version hash by querying github
	url := fmt.Sprintf("https://api.github.com/repos/flutter/engine/commits/%s", requiredEngineVersion)
	req, err := http.NewRequest("GET", os.ExpandEnv(url), nil)
	if err != nil {
		fmt.Printf("hover: Failed to create http request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", githubToken))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("hover: Failed to find engine version on github: %v\n", err)
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Printf("hover: Failed to read response body from github: %v\n", err)
		os.Exit(1)
	}

	// We define a struct to build JSON object from the response
	var apiResponse struct {
		Sha string `json:"sha"`
	}
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Printf("hover: Failed to unmarshall reply github: %v\n", err)
		os.Exit(1)
	}
	if apiResponse.Sha == "" {
		fmt.Printf("hover: Failed to fetch full sha for engine version %s from GitHub\n", requiredEngineVersion)
		os.Exit(1)
	}
	var requiredEngineVersionFullHash = apiResponse.Sha

	// TODO: support more arch's than x64?
	var platform = targetOS + "-x64"

	// Build the URL for downloading the correct engine
	var engineDownloadURL = fmt.Sprintf(targetedDomain+"/flutter_infra/flutter/%s/%s/", requiredEngineVersionFullHash, platform)
	switch targetOS {
	case "darwin":
		engineDownloadURL += "FlutterEmbedder.framework.zip"
	case "linux":
		engineDownloadURL += platform + "-embedder"
	case "windows":
		engineDownloadURL += platform + "-embedder.zip"
	default:
		fmt.Printf("hover: cannot run on %s, download engine not implemented.\n", runtime.GOOS)
		os.Exit(1)
	}

	icudtlDownloadURL := fmt.Sprintf(targetedDomain+"/flutter_infra/flutter/%s/%s/artifacts.zip", requiredEngineVersionFullHash, platform)

	dir, err := ioutil.TempDir("", "hover-engine-download")
	if err != nil {
		fmt.Printf("hover: Failed to create tmp dir for engine download: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	err = os.MkdirAll(dir, 0700)
	if err != nil {
		log.Fatal(err)
	}

	engineZipPath := filepath.Join(dir, "engine.zip")
	engineExtractPath := filepath.Join(dir, "engine")
	artifactsZipPath := filepath.Join(dir, "artifacts.zip")

	fmt.Printf("hover: Downloading engine for platform %s at version %s...\n", platform, requiredEngineVersion)
	err = downloadFile(engineZipPath, engineDownloadURL)
	if err != nil {
		fmt.Printf("hover: Failed to download engine: %v\n", err)
		os.Exit(1)
	}

	// TODO: make artifacts download a separate function, it doesn't need to be
	// downloaded with engine because it's OS independent.
	fmt.Printf("hover: Downloading artifacts at version %s...\n", requiredEngineVersion)
	err = downloadFile(artifactsZipPath, icudtlDownloadURL)
	if err != nil {
		fmt.Printf("hover: Failed to download artifacts: %v\n", err)
		os.Exit(1)
	}

	_, err = unzip(engineZipPath, engineExtractPath) // engineCachePath)
	if err != nil {
		log.Fatal(err)
	}

	artifactsCachePath := filepath.Join(engineCachePath, "artifacts")
	_, err = unzip(artifactsZipPath, artifactsCachePath) // filepath.Join(engineCachePath, "artifacts"))
	if err != nil {
		log.Fatal(err)
	}

	switch platform {
	case "darwin-x64":
		frameworkZipPath := filepath.Join(engineExtractPath, "FlutterEmbedder.framework.zip")
		frameworkDestPath := filepath.Join(engineCachePath, "FlutterEmbedder.framework")
		_, err = unzip(frameworkZipPath, frameworkDestPath)
		if err != nil {
			fmt.Printf("hover: Failed to unzip engine framework: %v\n", err)
			os.Exit(1)
		}

		createSymLink("A", frameworkDestPath+"/Versions/Current")
		createSymLink("Versions/Current/FlutterEmbedder", frameworkDestPath+"/FlutterEmbedder")
		createSymLink("Versions/Current/Headers", frameworkDestPath+"/Headers")
		createSymLink("Versions/Current/Modules", frameworkDestPath+"/Modules")
		createSymLink("Versions/Current/Resources", frameworkDestPath+"/Resources")

	case "linux-x64":
		err := moveFile(
			filepath.Join(engineExtractPath, "libflutter_engine.so"),
			filepath.Join(engineCachePath, "/libflutter_engine.so"),
		)
		if err != nil {
			fmt.Printf("hover: Failed to move downloaded libflutter_engine.so: %v\n", err)
			os.Exit(1)
		}

	case "windows-x64":
		err := moveFile(
			filepath.Join(engineExtractPath, "flutter_engine.dll"),
			filepath.Join(engineCachePath, "/flutter_engine.dll"),
		)
		if err != nil {
			fmt.Printf("hover: Failed to move downloaded flutter_engine.dll: %v\n", err)
			os.Exit(1)
		}
	}

	err = ioutil.WriteFile(cachedEngineVersionPath, []byte(requiredEngineVersion), 0664)
	if err != nil {
		fmt.Printf("hover: Failed to write version file: %v\n", err)
		os.Exit(1)
	}

	return
}

// ValidateOrUpdateEngine validates the engine we have cached matches the
// flutter version, or otherwise downloads a new engine. The returned path is
// that of the engine location.
func ValidateOrUpdateEngine(targetOS string) (engineCachePath string) {
	engineCachePath = ValidateOrUpdateEngineAtPath(targetOS, cachePath())
	return
}
