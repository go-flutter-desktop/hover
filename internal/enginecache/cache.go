package enginecache

import (
	"archive/zip"
	"encoding/json"
	"flag"
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

func createSymLink(from, to string) error {
	err := os.Remove(to)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove existing symlink")
	}

	err = os.Symlink(from, to)
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
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return nil, fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return nil, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return nil, err
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			if err != nil {
				return nil, errors.Wrap(err, "failed to copy file")
			}
		}
		filenames = append(filenames, fpath)
	}
	return filenames, nil
}

// Function to prind download percent completion
func printDownloadPercent(done chan chan struct{}, path string, expectedSize int64) {
	var completedCh chan struct{}
	for {
		file, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}

		fi, err := file.Stat()
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
	log.Printf("\033[2K\rDownload completed in %s\n", elapsed)

	return nil
}

// ValidateOrUpdateEngine validates the engine we have cached matches the
// flutter version, or otherwise downloads a new engine. The returned path is
// that of the engine location.
func ValidateOrUpdateEngine(targetOS string) (engineCachePath string) {
	engineCachePath = filepath.Join(cachePath(), "hover", "engine", targetOS)

	cachedEngineVersionPath := filepath.Join(engineCachePath, "version")
	cachedEngineVersionBytes, err := ioutil.ReadFile(cachedEngineVersionPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Failed to read cached engine version: %v\n", err)
		os.Exit(1)
	}
	cachedEngineVersion := string(cachedEngineVersionBytes)
	requiredEngineVersion := flutterRequiredEngineVersion()

	if cachedEngineVersion != "" {
		if cachedEngineVersion == requiredEngineVersion {
			fmt.Println("Using engine from cache")
			return
		}

		// Engine is outdated, we remove the old engine and continue to download
		// the new engine.
		err = os.RemoveAll(engineCachePath)
		if err != nil {
			fmt.Printf("Failed to remove outdated engine: %v\n", err)
			os.Exit(1)
		}
	}

	err = os.MkdirAll(engineCachePath, 0775)
	if err != nil {
		fmt.Printf("Failed to create engine cache directory: %v\n", err)
		os.Exit(1)
	}

	// TODO: move flag up the chain to cobra as flag and env var
	chinaPtr := flag.Bool("china", false, "Whether or not installation is in China")
	flag.Parse()

	// If flag china is passed, targeted domain is changed (China partially blocking google)
	var targetedDomain = ""
	if *chinaPtr {
		targetedDomain = "https://storage.flutter-io.cn"
	} else {
		targetedDomain = "https://storage.googleapis.com"
	}

	// Retrieve the full version hash by querying github
	url := fmt.Sprintf("https://api.github.com/search/commits?q=%s", requiredEngineVersion)
	req, err := http.NewRequest("GET", os.ExpandEnv(url), nil)
	if err != nil {
		fmt.Printf("Failed to create http request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Accept", "application/vnd.github.cloak-preview")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to find engine version on github: %v\n", err)
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Printf("Failed to read response body from github: %v\n", err)
		os.Exit(1)
	}

	// We define a struct to build JSON object from the response
	var hashResponse struct {
		Items []struct {
			Sha string `json:"sha"`
		} `json:"items"`
	}
	err = json.Unmarshal(body, &hashResponse)
	if err != nil {
		fmt.Printf("Failed to unmarshall reply github: %v\n", err)
		os.Exit(1)
	}
	var requiredEngineVersionFullHash = hashResponse.Items[0].Sha

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
		fmt.Printf("Glutter cannot run on %s, download engine not implemented.\n", runtime.GOOS)
		os.Exit(1)
	}

	icudtlDownloadURL := fmt.Sprintf(targetedDomain+"/flutter_infra/flutter/%s/%s/artifacts.zip", requiredEngineVersionFullHash, platform)

	dir, err := ioutil.TempDir("", "hover-engine-download")
	if err != nil {
		fmt.Printf("Failed to create tmp dir for engine download: %v\n", err)
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

	fmt.Printf("Downloading engine for platform %s at version %s...\n", platform, requiredEngineVersion)
	err = downloadFile(engineZipPath, engineDownloadURL)
	if err != nil {
		fmt.Printf("Failed to download engine: %v\n", err)
		os.Exit(1)
	}

	// TODO: make artifacts download a separate function, it doesn't need to be
	// downloaded with engine because it's OS independent.
	fmt.Printf("Downloading artifacts at version %s...\n", requiredEngineVersion)
	err = downloadFile(artifactsZipPath, icudtlDownloadURL)
	if err != nil {
		fmt.Printf("Failed to download artifacts: %v\n", err)
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
			fmt.Printf("Failed to unzip engine framework: %v\n", err)
			os.Exit(1)
		}

		// TODO: these symlinks are absolute and copied that way as well, this
		// doesn't work well for creating standalone applications. Investigate
		// what the symlinks are for, and how to make them relative so that an
		// application may be copied across machines/filesystems.
		createSymLink(frameworkDestPath+"/Versions/A", frameworkDestPath+"/Versions/Current")
		createSymLink(frameworkDestPath+"/Versions/Current/FlutterEmbedder", frameworkDestPath+"/FlutterEmbedder")
		createSymLink(frameworkDestPath+"/Versions/Current/Headers", frameworkDestPath+"/Headers")
		createSymLink(frameworkDestPath+"/Versions/Current/Modules", frameworkDestPath+"/Modules")
		createSymLink(frameworkDestPath+"/Versions/Current/Resources", frameworkDestPath+"/Resources")

	case "linux-x64":
		err := os.Rename(
			filepath.Join(engineExtractPath, "libflutter_engine.so"),
			filepath.Join(engineCachePath, "/libflutter_engine.so"),
		)
		if err != nil {
			fmt.Printf("Failed to move downloaded libflutter_engine.so: %v\n", err)
			os.Exit(1)
		}

	case "windows-x64":
		err := os.Rename(
			filepath.Join(engineExtractPath, "flutter_engine.dll"),
			filepath.Join(engineCachePath, "/flutter_engine.dll"),
		)
		if err != nil {
			fmt.Printf("Failed to move downloaded flutter_engine.dll: %v\n", err)
			os.Exit(1)
		}
	}

	err = ioutil.WriteFile(cachedEngineVersionPath, []byte(requiredEngineVersion), 0664)
	if err != nil {
		fmt.Printf("Failed to write version file: %v\n", err)
		os.Exit(1)
	}

	return
}
