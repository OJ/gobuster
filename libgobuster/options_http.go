package libgobuster

import (
	"crypto/tls"
	"net"
	"net/url"
	"time"
)

// BasicHTTPOptions defines only core http options
type BasicHTTPOptions struct {
	UserAgent        string
	Proxy            string
	NoTLSValidation  bool
	Timeout          time.Duration
	RetryOnTimeout   bool
	RetryAttempts    int
	TLSCertificate   *tls.Certificate
	TLSRenegotiation bool
	LocalAddr        *net.TCPAddr
}

// HTTPOptions is the struct to pass in all http options to Gobuster
type HTTPOptions struct {
	BasicHTTPOptions
	Password              string
	URL                   *url.URL
	Username              string
	Cookies               string
	Headers               []HTTPHeader
	NoCanonicalizeHeaders bool
	FollowRedirect        bool
	Method                string
	BodyOutputDir         string
}
