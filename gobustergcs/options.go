package gobustergcs

import (
	"github.com/OJ/gobuster/v3/libgobuster"
)

// OptionsGCS is the struct to hold all options for this plugin
type OptionsGCS struct {
	libgobuster.BasicHTTPOptions
	MaxFilesToList int
	ShowFiles      bool
}

// NewOptions returns a new initialized OptionsS3
func NewOptions() *OptionsGCS {
	return &OptionsGCS{}
}
