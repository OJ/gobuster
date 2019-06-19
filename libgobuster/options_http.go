package libgobuster

import (
	"time"
)

// OptionsHTTP is the struct to hold all options for common HTTP options
type OptionsHTTP struct {
	Password       string
	URL            string
	UserAgent      string
	Username       string
	Proxy          string
	Cookies        string
	Headers        []HTTPHeader
	Timeout        time.Duration
	FollowRedirect bool
	InsecureSSL    bool
}
