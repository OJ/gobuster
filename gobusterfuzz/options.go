package gobusterfuzz

import (
	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsFuzz is the struct to hold all options for this plugin
type OptionsFuzz struct {
	libgobuster.HTTPOptions
	Extensions                 string
	ExtensionsParsed           libgobuster.StringSet
	StatusCodes                string
	StatusCodesParsed          libgobuster.IntSet
	StatusCodesBlacklist       string
	StatusCodesBlacklistParsed libgobuster.IntSet
	UseSlash                   bool
	WildcardForced             bool
	IncludeLength              bool
	Expanded                   bool
	NoStatus                   bool
	DiscoverBackup             bool
}

// NewOptionsFuzz returns a new initialized OptionsFuzz
func NewOptionsFuzz() *OptionsFuzz {
	return &OptionsFuzz{
		StatusCodesParsed:          libgobuster.NewIntSet(),
		StatusCodesBlacklistParsed: libgobuster.NewIntSet(),
		ExtensionsParsed:           libgobuster.NewStringSet(),
	}
}
