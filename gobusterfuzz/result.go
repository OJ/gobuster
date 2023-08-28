package gobusterfuzz

import (
	"bytes"

	"github.com/fatih/color"
)

var (
	green = color.New(color.FgGreen).FprintfFunc()
)

// Result represents a single result
type Result struct {
	Word       string
	Path       string
	StatusCode int
	Size       int64
}

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	green(buf, "[Status=%d] [Length=%d] [Word=%s] %s\n", r.StatusCode, r.Size, r.Word, r.Path)

	s := buf.String()
	return s, nil
}
