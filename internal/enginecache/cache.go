package enginecache

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"

	"github.com/go-flutter-desktop/hover/internal/build"
	"github.com/go-flutter-desktop/hover/internal/darwinhacks"
	"github.com/go-flutter-desktop/hover/internal/log"
	"github.com/go-flutter-desktop/hover/internal/version"
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

		// Check for ZipSlip. More Infof: http://bit.ly/2MsjAWE
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
			log.Warnf("%v", err)
		}

		size := fi.Size()

		if size == 0 {
			size = 1
		}

		var percent = float64(size) / float64(expectedSize) * 100

		// We use '\033[2K\r' to avoid carriage return, it will print above previous.
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
	// // Printf download url in case user needs it.
	// log.Printf("Downloading file from\n '%s'\n to '%s'", url, filepath)

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
	log.Printf("\033[2K\rDownload completed in %.2fs", elapsed.Seconds())
	return nil
}

func EngineConfig(targetOS string, mode build.Mode) string {
	return fmt.Sprintf("%s-%s", targetOS, mode.Name)
}

//noinspection GoNameStartsWithPackageName
func EngineCachePath(targetOS, cachePath string, mode build.Mode) string {
	return filepath.Join(BaseEngineCachePath(cachePath), EngineConfig(targetOS, mode))
}

func BaseEngineCachePath(cachePath string) string {
	return filepath.Join(cachePath, "hover", "engine")
}

// ValidateOrUpdateEngine validates the engine we have cached matches the
// flutter version, or otherwise downloads a new engine. The engine cache
// location is set by the the user.
func ValidateOrUpdateEngine(targetOS, cachePath, requiredEngineVersion string, mode build.Mode) {
	engineCachePath := EngineCachePath(targetOS, cachePath, mode)

	if strings.Contains(engineCachePath, " ") {
		log.Errorf("Cannot save the engine to '%s', engine cache is not compatible with path containing spaces.", cachePath)
		log.Errorf("       Please run hover with a another engine cache path. Example:")
		log.Errorf("              %s", log.Au().Magenta("hover run --cache-path \"C:\\cache\""))
		log.Errorf("       The --cache-path flag will have to be provided to every build and run command.")
		os.Exit(1)
	}

	cachedEngineVersionPath := filepath.Join(engineCachePath, "version")
	cachedEngineVersionBytes, err := ioutil.ReadFile(cachedEngineVersionPath)
	if err != nil && !os.IsNotExist(err) {
		log.Errorf("Failed to read cached engine version: %v", err)
		os.Exit(1)
	}
	cachedEngineVersion := string(cachedEngineVersionBytes)
	if len(requiredEngineVersion) == 0 {
		requiredEngineVersion = version.FlutterRequiredEngineVersion()
	}

	if cachedEngineVersion == fmt.Sprintf("%s-%s", requiredEngineVersion, version.HoverVersion()) {
		log.Printf("Using engine from cache")
		return
	} else {
		// Engine is outdated, we remove the old engine and continue to download
		// the new engine.
		err = os.RemoveAll(engineCachePath)
		if err != nil {
			log.Errorf("Failed to remove outdated engine: %v", err)
			os.Exit(1)
		}
	}

	err = os.MkdirAll(engineCachePath, 0775)
	if err != nil {
		log.Errorf("Failed to create engine cache directory: %v", err)
		os.Exit(1)
	}

	dir, err := ioutil.TempDir("", "hover-engine-download")
	if err != nil {
		log.Errorf("Failed to create tmp dir for engine download: %v", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	err = os.MkdirAll(dir, 0700)
	if err != nil {
		log.Warnf("%v", err)
	}

	engineZipPath := filepath.Join(dir, "engine.zip")
	engineExtractPath := filepath.Join(dir, "engine")

	log.Printf("Downloading engine for platform %s at version %s...", EngineConfig(targetOS, mode), requiredEngineVersion)

	if mode == build.DebugMode {
		targetedDomain := "https://storage.googleapis.com"
		envURLFlutter := os.Getenv("FLUTTER_STORAGE_BASE_URL")
		if envURLFlutter != "" {
			targetedDomain = envURLFlutter
		}
		var engineDownloadURL = fmt.Sprintf(targetedDomain+"/flutter_infra/flutter/%s/%s-x64/", requiredEngineVersion, targetOS)
		switch targetOS {
		case "darwin":
			engineDownloadURL += "FlutterEmbedder.framework.zip"
		case "linux":
			engineDownloadURL += targetOS + "-x64-embedder"
		case "windows":
			engineDownloadURL += targetOS + "-x64-embedder.zip"
		default:
			log.Errorf("Cannot run on %s, download engine not implemented.", targetOS)
			os.Exit(1)
		}

		artifactsZipPath := filepath.Join(dir, "artifacts.zip")
		artifactsDownloadURL := fmt.Sprintf(targetedDomain+"/flutter_infra/flutter/%s/%s-x64/artifacts.zip", requiredEngineVersion, targetOS)

		err = downloadFile(engineZipPath, engineDownloadURL)
		if err != nil {
			log.Errorf("Failed to download engine: %v", err)
			os.Exit(1)
		}
		_, err = unzip(engineZipPath, engineExtractPath)
		if err != nil {
			log.Warnf("%v", err)
		}

		err = downloadFile(artifactsZipPath, artifactsDownloadURL)
		if err != nil {
			log.Errorf("Failed to download artifacts: %v", err)
			os.Exit(1)
		}
		_, err = unzip(artifactsZipPath, engineExtractPath)
		if err != nil {
			log.Warnf("%v", err)
		}
		if targetOS == "darwin" {
			frameworkZipPath := filepath.Join(engineExtractPath, "FlutterEmbedder.framework.zip")
			frameworkDestPath := filepath.Join(engineExtractPath, "FlutterEmbedder.framework")
			_, err = unzip(frameworkZipPath, frameworkDestPath)
			if err != nil {
				log.Errorf("Failed to unzip engine framework: %v", err)
				os.Exit(1)
			}
			createSymLink("A", frameworkDestPath+"/Versions/Current")
			createSymLink("Versions/Current/FlutterEmbedder", frameworkDestPath+"/FlutterEmbedder")
			createSymLink("Versions/Current/Headers", frameworkDestPath+"/Headers")
			createSymLink("Versions/Current/Modules", frameworkDestPath+"/Modules")
			createSymLink("Versions/Current/Resources", frameworkDestPath+"/Resources")
		}
	} else {
		file := ""
		switch targetOS {
		case "linux":
			file += "linux"
		case "darwin":
			file += "macosx"
		case "windows":
			file += "windows"
		}
		file += fmt.Sprintf("_x64-host_%s.zip", mode.Name)
		engineDownloadURL := fmt.Sprintf("https://github.com/flutter-rs/engine-builds/releases/download/f-%s/%s", requiredEngineVersion, file)

		err = downloadFile(engineZipPath, engineDownloadURL)
		if err != nil {
			log.Errorf("Failed to download engine: %v", err)
			log.Errorf("Engine builds are a bit delayed after they are published in flutter.")
			log.Errorf("You can either try again later or switch the flutter channel to beta, because these engines are more likely to be already built.")
			log.Errorf("To dig into the already built engines look at https://github.com/flutter-rs/engine-builds/releases and https://github.com/flutter-rs/engine-builds/actions")
			os.Exit(1)
		}
		_, err = unzip(engineZipPath, engineExtractPath)
		if err != nil {
			log.Warnf("%v", err)
		}

	}

	for _, engineFile := range build.EngineFiles(targetOS, mode) {
		err := copy.Copy(
			filepath.Join(engineExtractPath, engineFile),
			filepath.Join(engineCachePath, engineFile),
		)
		if err != nil {
			log.Errorf("Failed to copy downloaded %s: %v", engineFile, err)
			os.Exit(1)
		}
	}

	// Strip linux engine after download and not at every build
	if targetOS == "linux" {
		unstrippedEngineFile := filepath.Join(engineCachePath, build.EngineFiles(targetOS, mode)[0])
		err = exec.Command("strip", "-s", unstrippedEngineFile).Run()
		if err != nil {
			log.Errorf("Failed to strip %s: %v", unstrippedEngineFile, err)
			os.Exit(1)
		}
	}

	if targetOS == "darwin" && mode != build.DebugMode {
		darwinhacks.DyldHack(filepath.Join(engineCachePath, build.EngineFiles(targetOS, mode)[0]))
	}

	files := []string{
		"icudtl.dat",
	}
	if mode != build.DebugMode {
		files = append(
			files,
			"dart"+build.ExecutableExtension(targetOS),
			"gen_snapshot"+build.ExecutableExtension(targetOS),
			"gen",
			"flutter_patched_sdk",
		)
	}
	for _, file := range files {
		err = copy.Copy(
			filepath.Join(engineExtractPath, file),
			filepath.Join(engineCachePath, file),
		)
		if err != nil {
			log.Errorf("Failed to copy downloaded %s: %v", file, err)
			os.Exit(1)
		}
	}

	err = ioutil.WriteFile(cachedEngineVersionPath, []byte(fmt.Sprintf("%s-%s", requiredEngineVersion, version.HoverVersion())), 0664)
	if err != nil {
		log.Errorf("Failed to write version file: %v", err)
		os.Exit(1)
	}
}
