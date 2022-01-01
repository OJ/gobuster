package fuzz

import "github.com/OJ/gobuster/v3/lib"

// ErrWildcard is returned if a wildcard response is found
type ErrWildcard struct {
	url        string
	statusCode int
}

// GobusterFuzz is the main type to implement the interface
type GobusterFuzz struct {
	options    *OptionsFuzz
	globalopts *lib.Options
	http       *lib.HTTPClient
}

// OptionsFuzz is the struct to hold all options for this plugin
type OptionsFuzz struct {
	lib.HTTPOptions
	ExcludedStatusCodes       string
	ExcludedStatusCodesParsed lib.IntSet
	ExcludeLength             []int
}

// Result represents a single result
type Result struct {
	Verbose    bool
	Found      bool
	Path       string
	StatusCode int
	Size       int64
}
