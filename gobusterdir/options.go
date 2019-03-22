package gobusterdir

import (
	"time"

	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsDir is the struct to hold all options for this plugin
type OptionsDir struct {
	Extensions        string
	ExtensionsParsed  libgobuster.StringSet
	Password          string
	StatusCodes       string
	StatusCodesParsed libgobuster.IntSet
	URL               string
	UserAgent         string
	Username          string
	Proxy             string
	StringReplace     string
	Cookies           string
	Timeout           time.Duration
	FollowRedirect    bool
	IncludeLength     bool
	Expanded          bool
	NoStatus          bool
	InsecureSSL       bool
	UseSlash          bool
	IsWildcard        bool
	WildcardForced    bool
}

// NewOptionsDir returns a new initialized OptionsDir
func NewOptionsDir() *OptionsDir {
	return &OptionsDir{
		StatusCodesParsed: libgobuster.NewIntSet(),
		ExtensionsParsed:  libgobuster.NewStringSet(),
	}
}
