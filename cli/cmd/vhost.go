package cmd

import (
	"fmt"
	"log"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobustervhost"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

// nolint:gochecknoglobals
var cmdVhost *cobra.Command

func runVhost(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseVhostOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %w", err)
	}

	plugin, err := gobustervhost.NewGobusterVhost(globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobustervhost: %w", err)
	}

	if err := cli.Gobuster(mainContext, globalopts, plugin); err != nil {
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}

func parseVhostOptions() (*libgobuster.Options, *gobustervhost.OptionsVhost, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}
	var plugin gobustervhost.OptionsVhost

	httpOpts, err := parseCommonHTTPOptions(cmdVhost)
	if err != nil {
		return nil, nil, err
	}
	plugin.Password = httpOpts.Password
	plugin.URL = httpOpts.URL
	plugin.UserAgent = httpOpts.UserAgent
	plugin.Username = httpOpts.Username
	plugin.Proxy = httpOpts.Proxy
	plugin.Cookies = httpOpts.Cookies
	plugin.Timeout = httpOpts.Timeout
	plugin.FollowRedirect = httpOpts.FollowRedirect
	plugin.NoTLSValidation = httpOpts.NoTLSValidation
	plugin.Headers = httpOpts.Headers
	plugin.Method = httpOpts.Method

	plugin.AppendDomain, err = cmdVhost.Flags().GetBool("append-domain")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for append-domain: %w", err)
	}

	plugin.ExcludeLength, err = cmdVhost.Flags().GetIntSlice("exclude-length")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for excludelength: %w", err)
	}

	plugin.Domain, err = cmdVhost.Flags().GetString("domain")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for domain: %w", err)
	}

	return globalopts, &plugin, nil
}

// nolint:gochecknoinits
func init() {
	cmdVhost = &cobra.Command{
		Use:   "vhost",
		Short: "Uses VHOST enumeration mode (you most probably want to use the IP adress as the URL parameter",
		RunE:  runVhost,
	}
	if err := addCommonHTTPOptions(cmdVhost); err != nil {
		log.Fatalf("%v", err)
	}
	cmdVhost.Flags().BoolP("append-domain", "", false, "Append main domain from URL to words from wordlist. Otherwise the fully qualified domains need to be specified in the wordlist.")
	cmdVhost.Flags().IntSlice("exclude-length", []int{}, "exclude the following content length (completely ignores the status). Supply multiple times to exclude multiple sizes.")
	cmdVhost.Flags().String("domain", "", "the domain to append when using an IP address as URL. If left empty and you specify a domain based URL the hostname from the URL is extracted")

	cmdVhost.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdVhost)
}
