package gobustervhost

import (
	"time"
)

// OptionsVhost is the struct to hold all options for this plugin
type OptionsVhost struct {
	Password       string
	URL            string
	UserAgent      string
	Username       string
	Proxy          string
	Cookies        string
	Timeout        time.Duration
	InsecureSSL    bool
	FollowRedirect bool
}
