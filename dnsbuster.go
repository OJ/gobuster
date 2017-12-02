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
	wildcardIps, err := net.LookupHost(fmt.Sprintf("%s.%s", guid, cfg.url))
	if err == nil {
		cfg.isWildcard = true
		cfg.wildcardIps.addRange(wildcardIps)
		fmt.Println("[-] Wildcard DNS found. IP address(es): ", cfg.wildcardIps.string())
		if !cfg.wildcardForced {
			fmt.Println("[-] To force processing of Wildcard DNS, specify the '-fw' switch.")
		}
		return cfg.wildcardForced
	}

	if !cfg.quiet {
		// Provide a warning if the base domain doesn't resolve (in case of typo)
		_, err = net.LookupHost(cfg.url)
		if err != nil {
			// Not an error, just a warning. Eg. `yp.to` doesn't resolve, but `cr.py.to` does!
			fmt.Println("[-] Unable to validate base domain:", cfg.url)
		}
	}

	return true
}

func processDNS(cfg *config, word string, brc chan<- busterResult) {
	subdomain := word + "." + cfg.url
	ips, err := net.LookupHost(subdomain)

	if err == nil {
		if !cfg.isWildcard || !cfg.wildcardIps.containsAny(ips) {
			br := busterResult{
				entity: subdomain,
			}
			if cfg.showIPs {
				br.extra = strings.Join(ips, ", ")
			} else if cfg.showCNAME {
				cname, err := net.LookupCNAME(subdomain)
				if err == nil {
					br.extra = cname
				}
			}
			brc <- br
		}
	} else if cfg.verbose {
		br := busterResult{
			entity: subdomain,
			status: 404,
		}
		brc <- br
	}
}
