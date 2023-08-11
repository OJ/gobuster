package gobusterdir

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/fatih/color"
)

var (
	white  = color.New(color.FgWhite).FprintfFunc()
	yellow = color.New(color.FgYellow).FprintfFunc()
	green  = color.New(color.FgGreen).FprintfFunc()
	blue   = color.New(color.FgBlue).FprintfFunc()
	red    = color.New(color.FgRed).FprintfFunc()
	cyan   = color.New(color.FgCyan).FprintfFunc()
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
	buf := &bytes.Buffer{}

	// Prefix if we're in verbose mode
	if r.Verbose {
		if r.Found {
			if _, err := fmt.Fprintf(buf, "Found: "); err != nil {
				return "", err
			}
		} else {
			if _, err := fmt.Fprintf(buf, "Missed: "); err != nil {
				return "", err
			}
		}
	}

	if r.Expanded {
		if _, err := fmt.Fprintf(buf, "%s", r.URL); err != nil {
			return "", err
		}
	} else {
		if _, err := fmt.Fprintf(buf, "/"); err != nil {
			return "", err
		}
	}
	if _, err := fmt.Fprintf(buf, "%-20s", r.Path); err != nil {
		return "", err
	}

	if !r.NoStatus {
		color := white
		if r.StatusCode == 200 {
			color = green
		} else if r.StatusCode >= 300 && r.StatusCode < 400 {
			color = cyan
		} else if r.StatusCode >= 400 && r.StatusCode < 500 {
			color = yellow
		} else if r.StatusCode >= 500 && r.StatusCode < 600 {
			color = red
		}

		color(buf, " (Status: %d)", r.StatusCode)
	}

	if !r.HideLength {
		if _, err := fmt.Fprintf(buf, " [Size: %d]", r.Size); err != nil {
			return "", err
		}
	}

	location := r.Header.Get("Location")
	if location != "" {
		blue(buf, " [--> %s]", location)
	}

	if _, err := fmt.Fprintf(buf, "\n"); err != nil {
		return "", err
	}

	s := buf.String()

	return s, nil
}
