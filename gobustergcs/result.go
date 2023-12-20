package gobustergcs

import (
	"bytes"

	"github.com/fatih/color"
)

var (
	green = color.New(color.FgGreen).FprintfFunc()
)

// Result represents a single result
type Result struct {
	Found      bool
	BucketName string
	Status     string
}

// ResultToString converts the Result to its textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	c := green

	c(buf, "https://storage.googleapis.com/storage/v1/b/%s/o", r.BucketName)

	if r.Status != "" {
		c(buf, " [%s]", r.Status)
	}
	c(buf, "\n")

	str := buf.String()
	return str, nil
}
