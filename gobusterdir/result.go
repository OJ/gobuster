package gobusterdir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
)

var (
	white  = color.New(color.FgWhite).FprintfFunc()
	yellow = color.New(color.FgYellow).FprintfFunc()
	green  = color.New(color.FgGreen).FprintfFunc()
	blue   = color.New(color.FgBlue).FprintfFunc()
	red    = color.New(color.FgRed).FprintfFunc()
	cyan   = color.New(color.FgCyan).FprintfFunc()
)

// Result represents a single result
type Result struct {
	Path       string
	Header     http.Header
	StatusCode int
	Size       int64
}

// ResultToString converts the Result to its textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}
	if _, err := buf.WriteString(r.Path); err != nil {
		return "", err
	}

	if r.StatusCode >= 0 {
		textColor := white
		switch {
		case r.StatusCode == http.StatusOK:
			textColor = green
		case r.StatusCode >= 300 && r.StatusCode < 400:
			textColor = cyan
		case r.StatusCode >= 400 && r.StatusCode < 500:
			textColor = yellow
		case r.StatusCode >= 500 && r.StatusCode < 600:
			textColor = red
		}

		textColor(buf, " (Status: %d)", r.StatusCode)
	}

	if r.Size >= 0 {
		if _, err := fmt.Fprintf(buf, " [Size: %d]", r.Size); err != nil {
			return "", err
		}
	}

	location := r.Header.Get("Location")
	if location != "" {
		blue(buf, " [--> %s]", location)
	}

	if _, err := fmt.Fprintf(buf, "\n"); err != nil {
		return "", err
	}

	s := buf.String()

	return s, nil
}

// ResultToJSON converts the Result to JSON
func (r Result) ResultToJSON() ([]byte, error) {
	type JSONResult struct {
		Path       string      `json:"path"`
		StatusCode int         `json:"status_code"`
		Size       int64       `json:"size"`
		Location   string      `json:"location,omitempty"`
		Header     http.Header `json:"header,omitempty"`
	}

	jr := JSONResult{
		Path:       strings.TrimSpace(r.Path),
		StatusCode: r.StatusCode,
		Size:       r.Size,
		Location:   r.Header.Get("Location"),
		Header:     r.Header,
	}

	return json.Marshal(jr)
}
