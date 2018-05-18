package gobusterdns

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/OJ/gobuster/libgobuster"
	uuid "github.com/satori/go.uuid"
)

// SetupDNS is the setup implementation of gobusterdns
func SetupDNS(g *libgobuster.Gobuster) error {
	// Resolve a subdomain sthat probably shouldn't exist
	guid := uuid.Must(uuid.NewV4())
	wildcardIps, err := net.LookupHost(fmt.Sprintf("%s.%s", guid, g.Opts.URL))
	if err == nil {
		g.IsWildcard = true
		g.WildcardIps.AddRange(wildcardIps)
		log.Printf("[-] Wildcard DNS found. IP address(es): %s", g.WildcardIps.Stringify())
		if !g.Opts.WildcardForced {
			return fmt.Errorf("[-] To force processing of Wildcard DNS, specify the '-fw' switch.")
		}
	}

	if !g.Opts.Quiet {
		// Provide a warning if the base domain doesn't resolve (in case of typo)
		_, err = net.LookupHost(g.Opts.URL)
		if err != nil {
			// Not an error, just a warning. Eg. `yp.to` doesn't resolve, but `cr.py.to` does!
			log.Printf("[-] Unable to validate base domain: %s", g.Opts.URL)
		}
	}

	return nil
}

// ProcessDNSEntry is the process implementation of gobusterdns
func ProcessDNSEntry(g *libgobuster.Gobuster, word string) ([]libgobuster.Result, error) {
	subdomain := fmt.Sprintf("%s.%s", word, g.Opts.URL)
	ips, err := net.LookupHost(subdomain)
	var ret []libgobuster.Result
	if err == nil {
		if !g.IsWildcard || !g.WildcardIps.ContainsAny(ips) {
			result := libgobuster.Result{
				Entity: subdomain,
			}
			if g.Opts.ShowIPs {
				result.Extra = strings.Join(ips, ", ")
			} else if g.Opts.ShowCNAME {
				cname, err := net.LookupCNAME(subdomain)
				if err == nil {
					result.Extra = cname
				}
			}
			ret = append(ret, result)
		}
	} else if g.Opts.Verbose {
		ret = append(ret, libgobuster.Result{
			Entity: subdomain,
			Status: 404,
		})
	}
	return ret, nil
}

// DNSResultToString is the to string implementation of gobusterdns
func DNSResultToString(g *libgobuster.Gobuster, r *libgobuster.Result) (*string, error) {
	buf := &bytes.Buffer{}

	if r.Status == 404 {
		if _, err := fmt.Fprintf(buf, "Missing: %s\n", r.Entity); err != nil {
			return nil, err
		}
	} else if g.Opts.ShowIPs {
		if _, err := fmt.Fprintf(buf, "Found: %s [%s]\n", r.Entity, r.Extra); err != nil {
			return nil, err
		}
	} else if g.Opts.ShowCNAME {
		if _, err := fmt.Fprintf(buf, "Found: %s [%s]\n", r.Entity, r.Extra); err != nil {
			return nil, err
		}
	} else {
		if _, err := fmt.Fprintf(buf, "Found: %s\n", r.Entity); err != nil {
			return nil, err
		}
	}

	s := buf.String()
	return &s, nil
}
