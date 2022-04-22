package gobusterdns

import (
	"bytes"
	"fmt"
	"net/netip"
	"strings"
)

// Result represents a single result
type Result struct {
	ShowIPs   bool
	ShowCNAME bool
	Found     bool
	Subdomain string
	IPs       []netip.Addr
	CNAME     string
}

// ResultToString converts the Result to it's textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	if r.Found {
		if _, err := fmt.Fprintf(buf, "Found: "); err != nil {
			return "", err
		}
	} else {
		if _, err := fmt.Fprintf(buf, "Missed: "); err != nil {
			return "", err
		}
	}

	if r.ShowIPs && r.Found {
		ips := make([]string, len(r.IPs))
		for i := range r.IPs {
			ips[i] = r.IPs[i].String()
		}

		if _, err := fmt.Fprintf(buf, "%s [%s]\n", r.Subdomain, strings.Join(ips, ",")); err != nil {
			return "", err
		}
	} else if r.ShowCNAME && r.Found && r.CNAME != "" {
		if _, err := fmt.Fprintf(buf, "%s [%s]\n", r.Subdomain, r.CNAME); err != nil {
			return "", err
		}
	} else {
		if _, err := fmt.Fprintf(buf, "%s\n", r.Subdomain); err != nil {
			return "", err
		}
	}

	s := buf.String()
	return s, nil
}
