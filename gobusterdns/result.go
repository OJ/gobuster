package gobusterdns

import (
	"bytes"
	"net/netip"
	"strings"

	"github.com/fatih/color"
)

var (
	yellow = color.New(color.FgYellow).FprintfFunc()
	green  = color.New(color.FgGreen).FprintfFunc()
)

// Result represents a single result
type Result struct {
	ShowIPs   bool
	ShowCNAME bool
	Found     bool
	Subdomain string
	NoFQDN    bool
	IPs       []netip.Addr
	CNAME     string
}

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	c := green

	if !r.NoFQDN {
		r.Subdomain = strings.TrimSuffix(r.Subdomain, ".")
	}
	if r.Found {
		c(buf, "Found: ")
	} else {
		c = yellow
		c(buf, "Missed: ")
	}

	if r.ShowIPs && r.Found {
		ips := make([]string, len(r.IPs))
		for i := range r.IPs {
			ips[i] = r.IPs[i].String()
		}
		c(buf, "%s [%s]\n", r.Subdomain, strings.Join(ips, ","))
	} else if r.ShowCNAME && r.Found && r.CNAME != "" {
		c(buf, "%s [%s]\n", r.Subdomain, r.CNAME)
	} else {
		c(buf, "%s\n", r.Subdomain)
	}

	s := buf.String()
	return s, nil
}
