package gobusterdir

import (
	"bytes"
	"fmt"
	"net/http"
	"github.com/gookit/color"
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
	Colors     bool
	Header     http.Header
	StatusCode int
	Size       int64
}

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}
	colorize := fmt.Sprint

	// Apply coloring if flag set
	if r.Colors {
		green := color.FgGreen.Render
		yellow := color.FgYellow.Render
		red := color.FgRed.Render

		if !r.NoStatus {
			if r.StatusCode >= 200 && r.StatusCode < 300 {
				colorize = green
			} else if r.StatusCode >= 300 && r.StatusCode < 500 {
				colorize = yellow
			} else if r.StatusCode >= 500 && r.StatusCode <= 599 {
				colorize = red
			}
		}
	}

	// Prefix if we're in verbose mode
	if r.Verbose {
		if r.Found {
			if _, err := fmt.Fprintf(buf, colorize("Found: ")); err != nil {
				return "", err
			}
		} else {
			if _, err := fmt.Fprintf(buf, colorize("Missed: ")); err != nil {
				return "", err
			}
		}
	}

	if r.Expanded {
		s := colorize(fmt.Sprintf("%s", r.URL))
		if _, err := fmt.Fprintf(buf, s); err != nil {
			return "", err
		}
	} else {
		if _, err := fmt.Fprintf(buf, colorize("/")); err != nil {
			return "", err
		}
	}
	s := colorize(fmt.Sprintf("%-20s", r.Path))
	if _, err := fmt.Fprintf(buf, s); err != nil {
		return "", err
	}

	if !r.NoStatus {
		s = colorize(fmt.Sprintf(" (Status: %d)", r.StatusCode))
		if _, err := fmt.Fprintf(buf, s); err != nil {
			return "", err
		}
	}

	if !r.HideLength {
		s = colorize(fmt.Sprintf(" [Size: %d]", r.Size))
		if _, err := fmt.Fprintf(buf, s); err != nil {
			return "", err
		}
	}

	location := r.Header.Get("Location")
	if location != "" {
		s = colorize(fmt.Sprintf(" [--> %s]", location))
		if _, err := fmt.Fprintf(buf, s); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(buf, "\n"); err != nil {
		return "", err
	}

	s = buf.String()

	return s, nil
}
