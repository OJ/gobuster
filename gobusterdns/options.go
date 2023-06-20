package gobusterdns

import (
	"time"
)

// OptionsDNS holds all options for the dns plugin
type OptionsDNS struct {
	Domain         string
	ShowIPs        bool
	ShowCNAME      bool
	WildcardForced bool
	Resolver       string
	NoFQDN         bool
	Timeout        time.Duration
}

// NewOptionsDNS returns a new initialized OptionsDNS
func NewOptionsDNS() *OptionsDNS {
	return &OptionsDNS{}
}
