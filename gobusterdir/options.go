package gobusterdir

import (
	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsDir is the struct to hold all options for this plugin
type OptionsDir struct {
	libgobuster.OptionsHTTP
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
}

// NewOptionsDir returns a new initialized OptionsDir
func NewOptionsDir() *OptionsDir {
	return &OptionsDir{
		StatusCodesParsed:          libgobuster.NewIntSet(),
		StatusCodesBlacklistParsed: libgobuster.NewIntSet(),
		ExtensionsParsed:           libgobuster.NewStringSet(),
	}
}
