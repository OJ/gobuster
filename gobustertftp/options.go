package gobustertftp

import (
	"time"
)

// OptionsTFTP holds all options for the tftp plugin
type OptionsTFTP struct {
	Server  string
	Timeout time.Duration
}

// NewOptions returns a new initialized OptionsTFTP
func NewOptions() *OptionsTFTP {
	return &OptionsTFTP{}
}
