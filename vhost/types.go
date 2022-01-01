package vhost

import (
	"net/http"

	"github.com/OJ/gobuster/v3/lib"
)

// Result represents a single result
type Result struct {
	Found      bool
	Vhost      string
	StatusCode int
	Size       int64
	Header     http.Header
}

// GobusterVhost is the main type to implement the interface
type GobusterVhost struct {
	options    *OptionsVhost
	globalopts *lib.Options
	http       *lib.HTTPClient
	domain     string
	baseline1  []byte
	baseline2  []byte
}

// OptionsVhost is the struct to hold all options for this plugin
type OptionsVhost struct {
	lib.HTTPOptions
	AppendDomain  bool
	ExcludeLength []int
	Domain        string
}
