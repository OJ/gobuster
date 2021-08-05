package gobustervhost

import (
	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsVhost is the struct to hold all options for this plugin
type OptionsVhost struct {
	libgobuster.HTTPOptions
	AppendDomain               bool
	ExcludeLength              []int
	Domain                     string
	StatusCodes                string
	StatusCodesParsed          libgobuster.IntSet
	StatusCodesBlacklist       string
	StatusCodesBlacklistParsed libgobuster.IntSet
}

// NewOptionsDir returns a new initialized OptionsDir
func NewOptionsVhost() *OptionsVhost {
	return &OptionsVhost{
		StatusCodesParsed:          libgobuster.NewIntSet(),
		StatusCodesBlacklistParsed: libgobuster.NewIntSet(),
	}
}
