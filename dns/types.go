package dns

import (
	"net"
	"time"

	"github.com/OJ/gobuster/v3/lib"
)

// GobusterDNS is the main type to implement the interface
type GobusterDNS struct {
	resolver    *net.Resolver
	globalopts  *lib.Options
	options     *OptionsDNS
	isWildcard  bool
	wildcardIps lib.StringSet
}

// ErrWildcard is returned if a wildcard response is found
type ErrWildcard struct {
	wildcardIps lib.StringSet
}

// OptionsDNS holds all options for the dns plugin
type OptionsDNS struct {
	Domain         string
	ShowIPs        bool
	ShowCNAME      bool
	WildcardForced bool
	Resolver       string
	Timeout        time.Duration
}

// Result represents a single result
type Result struct {
	ShowIPs   bool
	ShowCNAME bool
	Found     bool
	Subdomain string
	IPs       []string
	CNAME     string
}
