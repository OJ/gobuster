package gobusters3

import (
	"bytes"
	"fmt"
)

// Result represents a single result
type Result struct {
	Found      bool
	BucketName string
	Status     string
}

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	if _, err := fmt.Fprintf(buf, "http://%s.s3.amazonaws.com/", r.BucketName); err != nil {
		return "", err
	}

	if r.Status != "" {
		if _, err := fmt.Fprintf(buf, " [%s]", r.Status); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(buf, "\n"); err != nil {
		return "", err
	}

	str := buf.String()
	return str, nil
}
