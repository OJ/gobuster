package dir

import (
	"net/http"

	"github.com/OJ/gobuster/v3/lib"
)

// Result represents a single result
type Result struct {
	URL        string
	Path       string
	Verbose    bool
	Expanded   bool
	NoStatus   bool
	HideLength bool
	Found      bool
	Header     http.Header
	StatusCode int
	Size       int64
}

// OptionsDir is the struct to hold all options for this plugin
type OptionsDir struct {
	lib.HTTPOptions
	Extensions                 string
	ExtensionsParsed           lib.StringSet
	StatusCodes                string
	StatusCodesParsed          lib.IntSet
	StatusCodesBlacklist       string
	StatusCodesBlacklistParsed lib.IntSet
	UseSlash                   bool
	HideLength                 bool
	Expanded                   bool
	NoStatus                   bool
	DiscoverBackup             bool
	ExcludeLength              []int
}

// ErrWildcard is returned if a wildcard response is found
type ErrWildcard struct {
	url        string
	statusCode int
	length     int64
}

// GobusterDir is the main type to implement the interface
type GobusterDir struct {
	options    *OptionsDir
	globalopts *lib.Options
	http       *lib.HTTPClient
}
