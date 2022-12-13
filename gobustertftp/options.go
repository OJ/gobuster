package gobustertftp

import (
	"time"
)

// OptionsTFTP holds all options for the tftp plugin
type OptionsTFTP struct {
	Server  string
	Timeout time.Duration
}

// NewOptionsTFTP returns a new initialized OptionsTFTP
func NewOptionsTFTP() *OptionsTFTP {
	return &OptionsTFTP{}
}
