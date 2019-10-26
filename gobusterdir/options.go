package gobusterdir

import (
	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsDir is the struct to hold all options for this plugin
type OptionsDir struct {
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
}

// NewOptionsDir returns a new initialized OptionsDir
func NewOptionsDir() *OptionsDir {
	httpOptions := libgobuster.NewHTTPOptions()
	return &OptionsDir{
		HTTPOptions:      *httpOptions,
		ExtensionsParsed: libgobuster.NewStringSet(),
	}
}
