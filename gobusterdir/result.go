package gobusterdir

import (
	"bytes"
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
		if _, err := fmt.Fprintf(buf, " (Status: %d)", r.StatusCode); err != nil {
			return "", err
		}
	}

	if !r.HideLength {
		if _, err := fmt.Fprintf(buf, " [Size: %d]", r.Size); err != nil {
			return "", err
		}
	}

	location := r.Header.Get("Location")
	if location != "" {
		if _, err := fmt.Fprintf(buf, " [--> %s]", location); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(buf, "\n"); err != nil {
		return "", err
	}

	s := buf.String()

	return s, nil
}
