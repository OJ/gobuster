package gobusterdns

import (
	"bytes"
	"net/netip"
	"strings"

	"github.com/fatih/color"
)

var (
	green = color.New(color.FgGreen).FprintfFunc()
)

// Result represents a single result
type Result struct {
	Subdomain string
	IPs       []netip.Addr
	CNAME     string
}

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	if len(r.IPs) > 0 {
		ips := make([]string, len(r.IPs))
		for i := range r.IPs {
			ips[i] = r.IPs[i].String()
		}
		green(buf, "Found: %s [%s]\n", r.Subdomain, strings.Join(ips, ","))
	}

	if r.CNAME != "" {
		green(buf, "Found CNAME: %s [%s]\n", r.Subdomain, r.CNAME)
	}

	s := buf.String()
	return s, nil
}
