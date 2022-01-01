package gobusterdir

import (
	"fmt"
	"net/http"
)

// Result represents a single result
type Result struct {
	URL        string
	Path       string
	Verbose    bool
	Expanded   bool
	NoStatus   bool
	HideLength bool
	Found      bool
	Header     http.Header
	StatusCode int
	Size       int64
}

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {

	var output string

	// Prefix if we're in verbose mode
	if r.Verbose {
		if r.Found {
			output = "Found: "
		} else {
			output = "Missed: "
		}
	}

	if r.Expanded {
		output += r.URL
	} else {
		output += "/"
	}

	output += fmt.Sprintf("%-20s", r.Path)

	if !r.NoStatus {
		output += fmt.Sprintf(" (Status: %d)", r.StatusCode)
	}

	if !r.HideLength {
		output += fmt.Sprintf(" [Size: %d]", r.Size)
	}

	location := r.Header.Get("Location")
	if location != "" {
		output += fmt.Sprintf(" [--> %s]", location)
	}

	output += "\n"

	return output, nil
}
