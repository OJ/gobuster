package cmd

import (
	"fmt"
	"log"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobustervhost"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

var cmdVhost *cobra.Command

func runVhost(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseVhostOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %w", err)
	}

	plugin, err := gobustervhost.NewGobusterVhost(mainContext, globalopts, pluginopts)
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
	plugin.AbsoluteVhost, err = cmdVhost.Flags().GetBool("absolute-vhost")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid argument for absolute-vhost: %w", err)
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

	return globalopts, &plugin, nil
}

func init() {
	cmdVhost = &cobra.Command{
		Use:   "vhost",
		Short: "Uses VHOST enumeration mode",
		RunE:  runVhost,
	}
	if err := addCommonHTTPOptions(cmdVhost); err != nil {
		log.Fatalf("%v", err)
	}

	cmdVhost.Flags().BoolP("absolute-vhost", "A", false, "vhost provided in wordlist are absolute")
	cmdVhost.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdVhost)
}
