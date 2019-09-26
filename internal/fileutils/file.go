package fileutils

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"

	rice "github.com/GeertJohan/go.rice"
	"github.com/go-flutter-desktop/hover/internal/log"
)

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

// CopyTemplate create file from a tempalte asset
func CopyTemplate(boxed, to string, assetsBox *rice.Box, templateData interface{}) {
	templateString, err := assetsBox.String(boxed)
	if err != nil {
		log.Errorf("Failed to find plugin template file: %v\n", err)
		os.Exit(1)
	}
	tmplFile, err := template.New("").Parse(templateString)
	if err != nil {
		log.Errorf("Failed to parse plugin template file: %v\n", err)
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
