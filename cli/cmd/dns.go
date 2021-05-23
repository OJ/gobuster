package cmd

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterdns"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

// nolint:gochecknoglobals
var cmdDNS *cobra.Command

func runDNS(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseDNSOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %w", err)
	}

	plugin, err := gobusterdns.NewGobusterDNS(globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobusterdns: %w", err)
	}

	if err := cli.Gobuster(mainContext, globalopts, plugin); err != nil {
		var wErr *gobusterdns.ErrWildcard
		if errors.As(err, &wErr) {
			return fmt.Errorf("%w. To force processing of Wildcard DNS, specify the '--wildcard' switch", wErr)
		}
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}

func parseDNSOptions() (*libgobuster.Options, *gobusterdns.OptionsDNS, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}
	plugin := gobusterdns.NewOptionsDNS()

	plugin.Domain, err = cmdDNS.Flags().GetString("domain")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for domain: %w", err)
	}

	plugin.ShowIPs, err = cmdDNS.Flags().GetBool("show-ips")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for show-ips: %w", err)
	}

	plugin.ShowCNAME, err = cmdDNS.Flags().GetBool("show-cname")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for show-cname: %w", err)
	}

	plugin.WildcardForced, err = cmdDNS.Flags().GetBool("wildcard")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for wildcard: %w", err)
	}

	plugin.Timeout, err = cmdDNS.Flags().GetDuration("timeout")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for timeout: %w", err)
	}

	plugin.Resolver, err = cmdDNS.Flags().GetString("resolver")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for resolver: %w", err)
	}

	if plugin.Resolver != "" && runtime.GOOS == "windows" {
		return nil, nil, fmt.Errorf("currently can not set custom dns resolver on windows. See https://golang.org/pkg/net/#hdr-Name_Resolution")
	}

	return globalopts, plugin, nil
}

// nolint:gochecknoinits
func init() {
	cmdDNS = &cobra.Command{
		Use:   "dns",
		Short: "Uses DNS subdomain enumeration mode",
		RunE:  runDNS,
	}

	cmdDNS.Flags().StringP("domain", "d", "", "The target domain")
	cmdDNS.Flags().BoolP("show-ips", "i", false, "Show IP addresses")
	cmdDNS.Flags().BoolP("show-cname", "c", false, "Show CNAME records (cannot be used with '-i' option)")
	cmdDNS.Flags().DurationP("timeout", "", time.Second, "DNS resolver timeout")
	cmdDNS.Flags().BoolP("wildcard", "", false, "Force continued operation when wildcard found")
	cmdDNS.Flags().StringP("resolver", "r", "", "Use custom DNS server (format server.com or server.com:port)")
	if err := cmdDNS.MarkFlagRequired("domain"); err != nil {
		log.Fatalf("error on marking flag as required: %v", err)
	}

	cmdDNS.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdDNS)
}
