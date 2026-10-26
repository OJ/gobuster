package gobusterdns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/netip"
	"strings"

	"github.com/fatih/color"
)

var green = color.New(color.FgGreen).FprintfFunc()

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
		green(buf, " %s", strings.Join(ips, ","))
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

// ResultToJSON converts the Result to JSON
func (r Result) ResultToJSON() ([]byte, error) {
	type JSONResult struct {
		Subdomain string   `json:"subdomain"`
		IPs       []string `json:"ips,omitempty"`
		CNAME     string   `json:"cname,omitempty"`
	}

	jr := JSONResult{
		Subdomain: r.Subdomain,
		CNAME:     r.CNAME,
	}

	if len(r.IPs) > 0 {
		jr.IPs = make([]string, len(r.IPs))
		for i := range r.IPs {
			jr.IPs[i] = r.IPs[i].String()
		}
	}

	return json.Marshal(jr)
}
