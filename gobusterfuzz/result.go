package gobusterfuzz

import (
	"bytes"
	"fmt"
)

// Result represents a single result
type Result struct {
	Verbose    bool
	Found      bool
	Path       string
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
	} else if r.Found {
		if _, err := fmt.Fprintf(buf, "Found: "); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(buf, "[Status=%d] [Length=%d] %s", r.StatusCode, r.Size, r.Path); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(buf, "\n"); err != nil {
		return "", err
	}

	s := buf.String()
	return s, nil
}
