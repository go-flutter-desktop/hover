package fileutils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	rice "github.com/GeertJohan/go.rice"

	"github.com/go-flutter-desktop/hover/internal/log"
)

// IsFileExists checks if a file exists and is not a directory
func IsFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// IsDirectory check if path exists and is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// RemoveLinesFromFile removes lines to a file if the text is present in the line
func RemoveLinesFromFile(filePath, text string) {
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Errorf("Failed to read file %s: %v\n", filePath, err)
		os.Exit(1)
	}

	lines := strings.Split(string(input), "\n")

	tmp := lines[:0]
	for _, line := range lines {
		if !strings.Contains(line, text) {
			tmp = append(tmp, line)
		}
	}
	output := strings.Join(tmp, "\n")
	err = ioutil.WriteFile(filePath, []byte(output), 0644)
	if err != nil {
		log.Errorf("Failed to write file %s: %v\n", filePath, err)
		os.Exit(1)
	}
}

// AddLineToFile appends a newLine to a file if the line isn't
// already present.
func AddLineToFile(filePath, newLine string) {
	f, err := os.OpenFile(filePath,
		os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Errorf("Failed to open file %s: %v\n", filePath, err)
		os.Exit(1)
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		log.Errorf("Failed to read file %s: %v\n", filePath, err)
		os.Exit(1)
	}
	lines := make(map[string]struct{})
	for _, w := range strings.Split(string(content), "\n") {
		lines[w] = struct{}{}
	}
	_, ok := lines[newLine]
	if ok {
		return
	}
	if _, err := f.WriteString(newLine + "\n"); err != nil {
		log.Errorf("Failed to append '%s' to the file (%s): %v\n", newLine, filePath, err)
		os.Exit(1)
	}
}

// CopyFile from one file to another
func CopyFile(src, to string) {
	in, err := os.Open(src)
	if err != nil {
		log.Errorf("Failed to read %s: %v\n", src, err)
		os.Exit(1)
	}
	defer in.Close()
	file, err := os.Create(to)
	if err != nil {
		log.Errorf("Failed to create %s: %v\n", to, err)
		os.Exit(1)
	}
	defer file.Close()

	_, err = io.Copy(file, in)
	if err != nil {
		log.Errorf("Failed to copy %s to %s: %v\n", src, to, err)
		os.Exit(1)
	}
}

// CopyDir copy files from one directory to another directory recursively
func CopyDir(src, dst string) {
	var err error
	var fds []os.FileInfo

	if !IsDirectory(src) {
		log.Errorf("Failed to copy directory, %s not a directory\n", src)
		os.Exit(1)
	}

	if err = os.MkdirAll(dst, 0755); err != nil {
		log.Errorf("Failed to copy directory %s to %s: %v\n", src, dst, err)
		os.Exit(1)
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		log.Errorf("Failed to list directory %s: %v\n", src, err)
		os.Exit(1)
	}

	for _, fd := range fds {
		srcPath := filepath.Join(src, fd.Name())
		dstPath := filepath.Join(dst, fd.Name())
		if fd.IsDir() {
			CopyDir(srcPath, dstPath)
		} else {
			CopyFile(srcPath, dstPath)
		}
	}
}

// CopyTemplateDir copy files from one directory to another directory recursively
// while executing all templates in files and file names
func CopyTemplateDir(boxed, to string, templateData interface{}) {
	var files []string
	err := filepath.Walk(boxed, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	files = files[1:]
	if err != nil {
		log.Errorf("Failed to list files in directory %s: %v\n", boxed, err)
		os.Exit(1)
	}
	for _, file := range files {
		newFile := filepath.Join(to, strings.Join(strings.Split(file, "")[len(boxed)+1:], ""))
		tmplFile, err := template.New("").Option("missingkey=error").Parse(newFile)
		if err != nil {
			log.Errorf("Failed to parse template string: %v\n", err)
			os.Exit(1)
		}
		var tmplBytes bytes.Buffer
		err = tmplFile.Execute(&tmplBytes, templateData)
		if err != nil {
			panic(err)
		}
		newFile = tmplBytes.String()
		fi, err := os.Stat(file)
		if err != nil {
			fmt.Println(err)
			return
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			err := os.MkdirAll(newFile, 0755)
			if err != nil {
				log.Errorf("Failed to create directory %s: %v\n", newFile, err)
				os.Exit(1)
			}
		case mode.IsRegular():
			if strings.HasSuffix(newFile, ".tmpl") {
				newFile = strings.TrimSuffix(newFile, ".tmpl")
			}
			ExecuteTemplateFromFile(file, newFile, templateData)
		}
	}
}

func executeTemplateFromString(templateString, to string, templateData interface{}) {
	tmplFile, err := template.New("").Option("missingkey=error").Parse(templateString)
	if err != nil {
		log.Errorf("Failed to parse template string: %v\n", err)
		os.Exit(1)
	}

	toFile, err := os.Create(to)
	if err != nil {
		log.Errorf("Failed to create '%s': %v\n", to, err)
		os.Exit(1)
	}
	defer toFile.Close()

	tmplFile.Execute(toFile, templateData)
}

// ExecuteTemplateFromFile create file from a template file
func ExecuteTemplateFromFile(boxed, to string, templateData interface{}) {
	templateString, err := ioutil.ReadFile(boxed)
	if err != nil {
		log.Errorf("Failed to find template file: %v\n", err)
		os.Exit(1)
	}
	executeTemplateFromString(string(templateString), to, templateData)
}

// ExecuteTemplateFromAssetsBox create file from a template asset
func ExecuteTemplateFromAssetsBox(boxed, to string, assetsBox *rice.Box, templateData interface{}) {
	templateString, err := assetsBox.String(boxed)
	if err != nil {
		log.Errorf("Failed to find template file: %v\n", err)
		os.Exit(1)
	}
	executeTemplateFromString(templateString, to, templateData)
}

// CopyAsset copies a file from asset
func CopyAsset(boxed, to string, assetsBox *rice.Box) {
	file, err := os.Create(to)
	if err != nil {
		log.Errorf("Failed to create %s: %v", to, err)
		os.Exit(1)
	}
	defer file.Close()
	boxedFile, err := assetsBox.Open(boxed)
	if err != nil {
		log.Errorf("Failed to find boxed file %s: %v", boxed, err)
		os.Exit(1)
	}
	defer boxedFile.Close()
	_, err = io.Copy(file, boxedFile)
	if err != nil {
		log.Errorf("Failed to write file %s: %v", to, err)
		os.Exit(1)
	}
}

// DownloadFile will download a url to a local file.
func DownloadFile(url string, filepath string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("Failed to download '%v': %v\n", url, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		log.Errorf("Failed to create file '%s': %v\n", filepath, err)
		os.Exit(1)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Errorf("Failed to write file '%s': %v\n", filepath, err)
		os.Exit(1)
	}
}
