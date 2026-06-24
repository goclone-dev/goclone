package parser

import (
	"net/url"
	"path"
)

// URLFilename returns the file name from a given url
func URLFilename(filename string) string {
	/*
		>>> https://tesla.com/main.css
		<<< main.css

		>>> https://dribbble.com/css/home.css
		<<< home.css

		>>> https://v2ex.com/assets/combo.css?t=1782285600
		<<< combo.css
	*/
	// Parse the URL so that any query string or fragment is dropped,
	// otherwise characters like "?" end up in the filename and are
	// invalid on platforms such as Windows.
	if u, err := url.Parse(filename); err == nil && u.Path != "" {
		return path.Base(u.Path)
	}
	return path.Base(filename)
}

// PathFilename returns the file name from a given path
func PathFilename(givenPath string) string {
	/*
		>>> /css/main.css
		<<< main.css

		>>> /js/googleanalytics.js
		<<< googleanalytics.js
	*/
	return path.Base(givenPath)
}
