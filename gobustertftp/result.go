package gobustertftp

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
	Filename string
	Found    bool
	Size     int64
}

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}
	c := green

	if r.Found {
		c(buf, "Found file %s", r.Filename)
		if r.Size > 0 {
			c(buf, " [%d]", r.Size)
		}
	}

	s := buf.String()
	return s, nil
}
