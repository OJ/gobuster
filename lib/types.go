package lib

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// GobusterPlugin is an interface which plugins must implement
type GobusterPlugin interface {
	Name() string
	PreRun(context.Context) error
	ProcessWord(context.Context, string, *Progress) error
	AdditionalWords(string) []string
	GetConfigString() (string, error)
}

// Gobuster is the main object when creating a new run
type Gobuster struct {
	Opts     *Options
	plugin   GobusterPlugin
	LogInfo  *log.Logger
	LogError *log.Logger
	Progress *Progress
}

// Progress holds all information regarding the running scan.
type Progress struct {
	requestsExpectedMutex *sync.RWMutex
	requestsExpected      int
	requestsCountMutex    *sync.RWMutex
	requestsIssued        int
	ResultChan            chan Result
	ErrorChan             chan error
}

// Options holds all options that can be passed to lib.
type Options struct {
	Threads        int
	Wordlist       string
	PatternFile    string
	Patterns       []string
	OutputFilename string
	NoStatus       bool
	NoProgress     bool
	NoError        bool
	Quiet          bool
	Verbose        bool
	Delay          time.Duration
}

// BasicHTTPOptions defines only core http options
type BasicHTTPOptions struct {
	UserAgent       string
	Proxy           string
	NoTLSValidation bool
	Timeout         time.Duration
	RetryOnTimeout  bool
	RetryAttempts   int
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

// HTTPHeader holds a single key value pair of a HTTP header
type HTTPHeader struct {
	Name  string
	Value string
}

// HTTPClient represents a http object
type HTTPClient struct {
	client           *http.Client
	userAgent        string
	defaultUserAgent string
	username         string
	password         string
	headers          []HTTPHeader
	cookies          string
	method           string
	host             string
}

// RequestOptions is used to pass options to a single individual request
type RequestOptions struct {
	Host       string
	Body       io.Reader
	ReturnBody bool
}

// IntSet is a set of Ints
type IntSet struct {
	Set map[int]bool
}

// StringSet is a set of Strings
type StringSet struct {
	Set map[string]bool
}
