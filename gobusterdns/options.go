package gobusterdns

import (
	"time"
)

// OptionsDNS holds all options for the dns plugin
type OptionsDNS struct {
	Domain         string
	CheckCNAME     bool
	WildcardForced bool
	Resolver       string
	Protocol       string
	NoFQDN         bool
	Timeout        time.Duration
}

// NewOptions returns a new initialized OptionsDNS
func NewOptions() *OptionsDNS {
	return &OptionsDNS{}
}
