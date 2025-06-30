package gobusterfuzz

import (
	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsFuzz is the struct to hold all options for this plugin
type OptionsFuzz struct {
	libgobuster.HTTPOptions
	ExcludedStatusCodes       string
	ExcludedStatusCodesParsed libgobuster.Set[int]
	ExcludeLength             string
	ExcludeLengthParsed       libgobuster.Set[int]
	RequestBody               string
}

// NewOptions returns a new initialized OptionsFuzz
func NewOptions() *OptionsFuzz {
	return &OptionsFuzz{
		ExcludedStatusCodesParsed: libgobuster.NewSet[int](),
		ExcludeLengthParsed:       libgobuster.NewSet[int](),
	}
}
