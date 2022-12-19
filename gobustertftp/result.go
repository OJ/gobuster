package gobustertftp

import (
	"bytes"
	"fmt"

	"github.com/fatih/color"
)

var (
	red   = color.New(color.FgRed).FprintfFunc()
	green = color.New(color.FgGreen).FprintfFunc()
)

// Result represents a single result
type Result struct {
	Filename     string
	Found        bool
	Size         int64
	ErrorMessage string
}

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	if r.Found {
		green(buf, "Found: ")
		if _, err := fmt.Fprintf(buf, "%s", r.Filename); err != nil {
			return "", err
		}
		if r.Size > 0 {
			if _, err := fmt.Fprintf(buf, " [%d]", r.Size); err != nil {
				return "", err
			}
		}
	} else {
		red(buf, "Missed: ")
		if _, err := fmt.Fprintf(buf, "%s - %s", r.Filename, r.ErrorMessage); err != nil {
			return "", err
		}
	}

	s := buf.String()
	return s, nil
}
