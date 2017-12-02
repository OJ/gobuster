// DNS buster

package main

import (
	"fmt"
	"net"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func setupDNS(cfg *config) bool {
	// Resolve a subdomain that probably shouldn't exist
	guid := uuid.NewV4()
	wildcardIps, err := net.LookupHost(fmt.Sprintf("%s.%s", guid, cfg.Url))
	if err == nil {
		cfg.IsWildcard = true
		cfg.WildcardIps.addRange(wildcardIps)
		fmt.Println("[-] Wildcard DNS found. IP address(es): ", cfg.WildcardIps.string())
		if !cfg.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard DNS, specify the '-fw' switch.")
		}
		return cfg.WildcardForced
	}

	if !cfg.Quiet {
		// Provide a warning if the base domain doesn't resolve (in case of typo)
		_, err = net.LookupHost(cfg.Url)
		if err != nil {
			// Not an error, just a warning. Eg. `yp.to` doesn't resolve, but `cr.py.to` does!
			fmt.Println("[-] Unable to validate base domain:", cfg.Url)
		}
	}

	return true
}

func processDNS(cfg *config, word string, brc chan<- busterResult) {
	subdomain := word + "." + cfg.Url
	ips, err := net.LookupHost(subdomain)

	if err == nil {
		if !cfg.IsWildcard || !cfg.WildcardIps.containsAny(ips) {
			br := busterResult{
				Entity: subdomain,
			}
			if cfg.ShowIPs {
				br.Extra = strings.Join(ips, ", ")
			} else if cfg.ShowCNAME {
				cname, err := net.LookupCNAME(subdomain)
				if err == nil {
					br.Extra = cname
				}
			}
			brc <- br
		}
	} else if cfg.Verbose {
		br := busterResult{
			Entity: subdomain,
			Status: 404,
		}
		brc <- br
	}
}
