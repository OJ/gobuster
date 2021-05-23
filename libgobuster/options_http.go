package libgobuster

import (
	"time"
)

// BasicHTTPOptions defines only core http options
type BasicHTTPOptions struct {
	UserAgent       string
	Proxy           string
	NoTLSValidation bool
	Timeout         time.Duration
}

// HTTPOptions is the struct to pass in all http options to Gobuster
type HTTPOptions struct {
	BasicHTTPOptions
	Password       string
	URL            string
	Username       string
	Cookies        string
	Headers        []HTTPHeader
	FollowRedirect bool
	Method         string
}
