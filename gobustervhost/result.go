package gobustervhost

import (
	"bytes"
	"fmt"
	"net/http"
)

// Result represents a single result
type Result struct {
	Found      bool
	Vhost      string
	StatusCode int
	Size       int64
	Header     http.Header
}

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
