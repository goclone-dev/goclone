package html

import (
	"os"

	"github.com/yosssi/gohtml"
)

// FormatHTML will formart any given string of HTML
func FormatHTML(filePath string) {
	// TODO: Implement
	dat, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	data := gohtml.Format(string(dat))
	b := []byte(data)
	error := os.WriteFile(filePath, b, 0777)
	// handle this error
	if error != nil {
		// print it out
		panic(error)
	}
}
