package gobustergcs

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
)

var green = color.New(color.FgGreen).FprintfFunc()

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

// ResultToJSON converts the Result to JSON
func (r Result) ResultToJSON() ([]byte, error) {
	type JSONResult struct {
		Found      bool   `json:"found"`
		BucketName string `json:"bucket_name"`
		Status     string `json:"status,omitempty"`
		URL        string `json:"url"`
	}

	jr := JSONResult{
		Found:      r.Found,
		BucketName: r.BucketName,
		Status:     r.Status,
		URL:        fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o", r.BucketName),
	}

	return json.Marshal(jr)
}
