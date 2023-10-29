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
	ExcludeStatus       string
	ExcludeStatusParsed libgobuster.Set[int]
	Domain              string
}

// NewOptions returns a new initialized OptionsVhost
func NewOptions() *OptionsVhost {
	return &OptionsVhost{
		ExcludeLengthParsed: libgobuster.NewSet[int](),
		ExcludeStatusParsed: libgobuster.NewSet[int](),
	}
}
