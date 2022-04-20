package gobustervhost

import (
	"fmt"
	"net/http"

	"github.com/fatih/color"
)

var (
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
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
	statusText := yellow("Missed")
	if r.Found {
		statusText = green("Found")
	}

	statusCode := yellow(fmt.Sprintf("Status: %d", r.StatusCode))
	if r.StatusCode == 200 {
		statusCode = green(fmt.Sprintf("Status: %d", r.StatusCode))
	} else if r.StatusCode >= 500 && r.StatusCode < 600 {
		statusCode = red(fmt.Sprintf("Status: %d", r.StatusCode))
	}

	location := r.Header.Get("Location")
	locationString := ""
	if location != "" {
		locationString = blue(fmt.Sprintf(" [--> %s]", location))
	}

	return fmt.Sprintf("%s: %s %s [Size: %d]%s\n", statusText, r.Vhost, statusCode, r.Size, locationString), nil
}
