package gobustervhost

import (
	"fmt"
	"net/http"

	"github.com/fatih/color"
)

var (
	white  = color.New(color.FgWhite).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
)

// Result represents a single result
type Result struct {
	Vhost      string
	StatusCode int
	Size       int64
	Header     http.Header
}

// ResultToString converts the Result to its textual representation
func (r Result) ResultToString() (string, error) {
	statusCodeColor := white
	if r.StatusCode == 200 {
		statusCodeColor = green
	} else if r.StatusCode >= 300 && r.StatusCode < 400 {
		statusCodeColor = cyan
	} else if r.StatusCode >= 400 && r.StatusCode < 500 {
		statusCodeColor = yellow
	} else if r.StatusCode >= 500 && r.StatusCode < 600 {
		statusCodeColor = red
	}

	statusCode := statusCodeColor(fmt.Sprintf("Status: %d", r.StatusCode))

	location := r.Header.Get("Location")
	locationString := ""
	if location != "" {
		locationString = blue(fmt.Sprintf(" [--> %s]", location))
	}

	return fmt.Sprintf("%s %s [Size: %d]%s\n", r.Vhost, statusCode, r.Size, locationString), nil
}
