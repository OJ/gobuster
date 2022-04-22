package gobustergcs

import (
	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsGCS is the struct to hold all options for this plugin
type OptionsGCS struct {
	libgobuster.BasicHTTPOptions
	MaxFilesToList int
}

// NewOptionsGCS returns a new initialized OptionsS3
func NewOptionsGCS() *OptionsGCS {
	return &OptionsGCS{}
}
