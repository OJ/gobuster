package gobustervhost

import (
	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsVhost is the struct to hold all options for this plugin
type OptionsVhost struct {
	libgobuster.HTTPOptions
	StatusCodes                string
	StatusCodesParsed          libgobuster.Set[int]
	StatusCodesBlacklist       string
	StatusCodesBlacklistParsed libgobuster.Set[int]
	AppendDomain        bool
	ExcludeLength       string
	ExcludeLengthParsed libgobuster.Set[int]
	Domain              string
}

// NewOptionsVhost returns a new initialized OptionsVhost
func NewOptionsVhost() *OptionsVhost {
	return &OptionsVhost{
		StatusCodesParsed:          libgobuster.NewSet[int](),
		StatusCodesBlacklistParsed: libgobuster.NewSet[int](),
		ExcludeLengthParsed: libgobuster.NewSet[int](),
	}
}
