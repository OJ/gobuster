package libgobuster

import (
	"fmt"
	"net"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func SetupDns(s *State) bool {
	// Resolve a subdomain that probably shouldn't exist
	guid := uuid.Must(uuid.NewV4())
	wildcardIps, err := net.LookupHost(fmt.Sprintf("%s.%s", guid, s.URL))
	if err == nil {
		s.IsWildcard = true
		s.WildcardIps.AddRange(wildcardIps)
		fmt.Println("[-] Wildcard DNS found. IP address(es): ", s.WildcardIps.Stringify())
		if !s.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard DNS, specify the '-fw' switch.")
		}
		return s.WildcardForced
	}

	if !s.Quiet {
		// Provide a warning if the base domain doesn't resolve (in case of typo)
		_, err = net.LookupHost(s.URL)
		if err != nil {
			// Not an error, just a warning. Eg. `yp.to` doesn't resolve, but `cr.py.to` does!
			fmt.Println("[-] Unable to validate base domain:", s.URL)
		}
	}

	return true
}

func ProcessDnsEntry(s *State, word string, resultChan chan<- Result) {
	subdomain := fmt.Sprintf("%s.%s", word, s.URL)
	ips, err := net.LookupHost(subdomain)

	if err == nil {
		if !s.IsWildcard || !s.WildcardIps.ContainsAny(ips) {
			result := Result{
				Entity: subdomain,
			}
			if s.ShowIPs {
				result.Extra = strings.Join(ips, ", ")
			} else if s.ShowCNAME {
				cname, err := net.LookupCNAME(subdomain)
				if err == nil {
					result.Extra = cname
				}
			}
			resultChan <- result
		}
	} else if s.Verbose {
		result := Result{
			Entity: subdomain,
			Status: 404,
		}
		resultChan <- result
	}
}

func PrintDnsResult(s *State, r *Result) {
	var output string
	if r.Status == 404 {
		output = fmt.Sprintf("Missing: %s\n", r.Entity)
	} else if s.ShowIPs {
		output = fmt.Sprintf("Found: %s [%s]\n", r.Entity, r.Extra)
	} else if s.ShowCNAME {
		output = fmt.Sprintf("Found: %s [%s]\n", r.Entity, r.Extra)
	} else {
		output = fmt.Sprintf("Found: %s\n", r.Entity)
	}
	fmt.Printf("%s", output)

	if s.OutputFile != nil {
		WriteToFile(output, s)
	}
}
