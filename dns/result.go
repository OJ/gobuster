package dns

import (
	"bytes"
	"fmt"
	"strings"
)

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
		if _, err := fmt.Fprintf(buf, "%s [%s]\n", r.Subdomain, strings.Join(r.IPs, ",")); err != nil {
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
