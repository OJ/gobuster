package vhost

import (
	"bytes"
	"fmt"
)

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	statusText := "Missed"
	if r.Found {
		statusText = "Found"
	}

	if _, err := fmt.Fprintf(buf, "%s: %s (Status: %d) [Size: %d]\n", statusText, r.Vhost, r.StatusCode, r.Size); err != nil {
		return "", err
	}

	s := buf.String()
	return s, nil
}
