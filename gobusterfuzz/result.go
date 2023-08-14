package gobusterfuzz

import (
	"bytes"

	"github.com/fatih/color"
)

var (
	yellow = color.New(color.FgYellow).FprintfFunc()
	green  = color.New(color.FgGreen).FprintfFunc()
)

// Result represents a single result
type Result struct {
	Word       string
	Verbose    bool
	Found      bool
	Path       string
	StatusCode int
	Size       int64
}

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	c := green

	// Prefix if we're in verbose mode
	if r.Verbose {
		if r.Found {
			c(buf, "Found: ")
		} else {
			c = yellow
			c(buf, "Missed: ")
		}
	} else if r.Found {
		c(buf, "Found: ")
	}

	c(buf, "[Status=%d] [Length=%d] [Word=%s] %s", r.StatusCode, r.Size, r.Word, r.Path)
	c(buf, "\n")

	s := buf.String()
	return s, nil
}
