package gobusterdns

import (
	"bytes"
	"fmt"
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

// ResultToString converts the Result to its textual representation
func (r Result) ResultToString() (string, error) {
	buf := &bytes.Buffer{}

	if _, err := fmt.Fprintf(buf, "%s", r.Subdomain); err != nil {
		return "", err
	}

	if len(r.IPs) > 0 {
		ips := make([]string, len(r.IPs))
		for i := range r.IPs {
			ips[i] = r.IPs[i].String()
		}
		green(buf, " IPs: %s", strings.Join(ips, ","))
	}

	if r.CNAME != "" {
		green(buf, " CNAME: %s", r.CNAME)
	}

	if _, err := fmt.Fprintf(buf, "\n"); err != nil {
		return "", err
	}

	s := buf.String()
	return s, nil
}
