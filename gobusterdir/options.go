package gobusterdir

import (
	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsDir is the struct to hold all options for this plugin
type OptionsDir struct {
	libgobuster.HTTPOptions
	Extensions                 string
	ExtensionsParsed           libgobuster.Set[string]
	ExtensionsFile             string
	StatusCodes                string
	StatusCodesParsed          libgobuster.Set[int]
	StatusCodesBlacklist       string
	StatusCodesBlacklistParsed libgobuster.Set[int]
	UseSlash                   bool
	HideLength                 bool
	Expanded                   bool
	NoStatus                   bool
	DiscoverBackup             bool
	ExcludeLength              []int
}

// NewOptionsDir returns a new initialized OptionsDir
func NewOptionsDir() *OptionsDir {
	return &OptionsDir{
		StatusCodesParsed:          libgobuster.NewSet[int](),
		StatusCodesBlacklistParsed: libgobuster.NewSet[int](),
		ExtensionsParsed:           libgobuster.NewSet[string](),
	}
}
