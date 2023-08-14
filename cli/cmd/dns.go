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

	log := libgobuster.NewLogger(globalopts.Debug)
	if err := cli.Gobuster(mainContext, globalopts, plugin, log); err != nil {
		var wErr *gobusterdns.ErrWildcard
		if errors.As(err, &wErr) {
			return fmt.Errorf("%w. To force processing of Wildcard DNS, specify the '--wildcard' switch", wErr)
		}
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}

func parseDNSOptions() (*libgobuster.Options, *gobusterdns.OptionsDNS, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}
	pluginOpts := gobusterdns.NewOptionsDNS()

	pluginOpts.Domain, err = cmdDNS.Flags().GetString("domain")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for domain: %w", err)
	}

	pluginOpts.ShowIPs, err = cmdDNS.Flags().GetBool("show-ips")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for show-ips: %w", err)
	}

	pluginOpts.ShowCNAME, err = cmdDNS.Flags().GetBool("show-cname")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for show-cname: %w", err)
	}

	pluginOpts.WildcardForced, err = cmdDNS.Flags().GetBool("wildcard")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for wildcard: %w", err)
	}

	pluginOpts.Timeout, err = cmdDNS.Flags().GetDuration("timeout")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for timeout: %w", err)
	}

	pluginOpts.Resolver, err = cmdDNS.Flags().GetString("resolver")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for resolver: %w", err)
	}

	pluginOpts.NoFQDN, err = cmdDNS.Flags().GetBool("no-fqdn")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for no-fqdn: %w", err)
	}

	if pluginOpts.Resolver != "" && runtime.GOOS == "windows" {
		return nil, nil, fmt.Errorf("currently can not set custom dns resolver on windows. See https://golang.org/pkg/net/#hdr-Name_Resolution")
	}

	return globalopts, pluginOpts, nil
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
	cmdDNS.Flags().BoolP("no-fqdn", "", false, "Do not automatically add a trailing dot to the domain, so the resolver uses the DNS search domain")
	cmdDNS.Flags().StringP("resolver", "r", "", "Use custom DNS server (format server.com or server.com:port)")
	if err := cmdDNS.MarkFlagRequired("domain"); err != nil {
		log.Fatalf("error on marking flag as required: %v", err)
	}

	cmdDNS.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdDNS)
}
