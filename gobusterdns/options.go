package gobusterdns

import (
	"time"
)

// OptionsDNS holds all options for the dns plugin
type OptionsDNS struct {
	Domain         string
	ShowIPs        bool
	CheckCNAME     bool
	WildcardForced bool
	Resolver       string
	NoFQDN         bool
	Timeout        time.Duration
}

// NewOptions returns a new initialized OptionsDNS
func NewOptions() *OptionsDNS {
	return &OptionsDNS{}
}
