package gobustervhost

import (
	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsVhost is the struct to hold all options for this plugin
type OptionsVhost struct {
	libgobuster.HTTPOptions
	AppendDomain        bool
	ExcludeLength       string
	ExcludeLengthParsed libgobuster.Set[int]
	Domain              string
}

// NewOptionsVhost returns a new initialized OptionsVhost
func NewOptionsVhost() *OptionsVhost {
	return &OptionsVhost{
		ExcludeLengthParsed: libgobuster.NewSet[int](),
	}
}
