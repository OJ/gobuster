package gobusterfuzz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fatih/color"
)

var (
	green = color.New(color.FgGreen).FprintfFunc()
	blue  = color.New(color.FgBlue).FprintfFunc()
)

// Result represents a single result
type Result struct {
	Word       string
	Path       string
	StatusCode int
	Size       int64
	Header     http.Header
}

// ResultToString converts the Result to its textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	green(buf, "[Status=%d] [Length=%d] [Word=%s] %s", r.StatusCode, r.Size, r.Word, r.Path)

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
		Word       string      `json:"word"`
		Path       string      `json:"path"`
		StatusCode int         `json:"status_code"`
		Size       int64       `json:"size"`
		Location   string      `json:"location,omitempty"`
		Header     http.Header `json:"header,omitempty"`
	}

	jr := JSONResult{
		Word:       r.Word,
		Path:       r.Path,
		StatusCode: r.StatusCode,
		Size:       r.Size,
		Location:   r.Header.Get("Location"),
		Header:     r.Header,
	}

	return json.Marshal(jr)
}
