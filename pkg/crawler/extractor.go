package crawler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/goclone-dev/goclone/pkg/parser"
)

// file extension map for directing files to their proper directory in O(1) time
var (
	extensionDir = map[string]string{
		".css":  "css",
		".js":   "js",
		".jpg":  "imgs",
		".jpeg": "imgs",
		".gif":  "imgs",
		".png":  "imgs",
		".svg":  "imgs",
	}
)

// Extractor visits a link determines if its a page or sublink
// downloads the contents to a correct directory in project folder
// TODO add functionality for determining if page or sublink
func Extractor(link string, projectPath string) {
	fmt.Println("Extracting --> ", link)

	// get the html body
	resp, err := http.Get(link)
	if err != nil {
		panic(err)
	}

	// Closure
	defer resp.Body.Close()

	// Get the original filename from the URL
	base := parser.URLFilename(link)
	// Get the clean extension
	ext := parser.URLExtension(link)

	// checks if there was a valid extension
	if ext != "" {
		// checks if that extension has a directory path name associated with it
		// from the extensionDir map
		dirPath := extensionDir[ext]
		if dirPath != "" {
			// If extension and path are valid pass to writeFileToPath
			writeFileToPath(projectPath, base, dirPath, resp)
		}
	}
}

func writeFileToPath(projectPath, filename, fileDir string, resp *http.Response) {
	// Create the full path
	fullPath := filepath.Join(projectPath, fileDir, filename)

	// Create the directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(fullPath), 0777)
	if err != nil {
		panic(err)
	}

	// Open the file for writing
	f, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Read and write the file contents
	htmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	f.Write(htmlData)
}
