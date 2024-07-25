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

	log := libgobuster.NewLogger(globalopts.Debug)
	if err := cli.Gobuster(mainContext, globalopts, plugin, log); err != nil {
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}

func parseVhostOptions() (*libgobuster.Options, *gobustervhost.OptionsVhost, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}

	pluginOpts := gobustervhost.NewOptionsVhost()

	httpOpts, err := parseCommonHTTPOptions(cmdVhost)
	if err != nil {
		return nil, nil, err
	}
	pluginOpts.Password = httpOpts.Password
	pluginOpts.URL = httpOpts.URL
	pluginOpts.UserAgent = httpOpts.UserAgent
	pluginOpts.Username = httpOpts.Username
	pluginOpts.Proxy = httpOpts.Proxy
	pluginOpts.Cookies = httpOpts.Cookies
	pluginOpts.Timeout = httpOpts.Timeout
	pluginOpts.FollowRedirect = httpOpts.FollowRedirect
	pluginOpts.NoTLSValidation = httpOpts.NoTLSValidation
	pluginOpts.Headers = httpOpts.Headers
	pluginOpts.Method = httpOpts.Method
	pluginOpts.RetryOnTimeout = httpOpts.RetryOnTimeout
	pluginOpts.RetryAttempts = httpOpts.RetryAttempts
	pluginOpts.TLSCertificate = httpOpts.TLSCertificate
	pluginOpts.NoCanonicalizeHeaders = httpOpts.NoCanonicalizeHeaders

	pluginOpts.AppendDomain, err = cmdVhost.Flags().GetBool("append-domain")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for append-domain: %w", err)
	}

	pluginOpts.ExcludeLength, err = cmdVhost.Flags().GetString("exclude-length")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for exclude-length: %w", err)
	}
	ret, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.ExcludeLength)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for exclude-length: %w", err)
	}
	pluginOpts.ExcludeLengthParsed = ret

	pluginOpts.Domain, err = cmdVhost.Flags().GetString("domain")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for domain: %w", err)
	}

	// parse normal status codes
	pluginOpts.StatusCodes, err = cmdVhost.Flags().GetString("status-codes")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for status-codes: %w", err)
	}
	ret2, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.StatusCodes)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for status-codes: %w", err)
	}
	pluginOpts.StatusCodesParsed = ret2

	// blacklist will override the normal status codes
	pluginOpts.StatusCodesBlacklist, err = cmdVhost.Flags().GetString("status-codes-blacklist")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for status-codes-blacklist: %w", err)
	}
	ret3, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.StatusCodesBlacklist)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for status-codes-blacklist: %w", err)
	}
	pluginOpts.StatusCodesBlacklistParsed = ret3

	if pluginOpts.StatusCodes != "" && pluginOpts.StatusCodesBlacklist != "" {
		return nil, nil, fmt.Errorf("status-codes (%q) and status-codes-blacklist (%q) are both set - please set only one. status-codes-blacklist is set by default so you might want to disable it by supplying an empty string.",
			pluginOpts.StatusCodes, pluginOpts.StatusCodesBlacklist)
	}

	if pluginOpts.StatusCodes == "" && pluginOpts.StatusCodesBlacklist == "" {
		return nil, nil, fmt.Errorf("status-codes and status-codes-blacklist are both not set, please set one")
	}

	return globalopts, pluginOpts, nil
}

// nolint:gochecknoinits
func init() {
	cmdVhost = &cobra.Command{
		Use:   "vhost",
		Short: "Uses VHOST enumeration mode (you most probably want to use the IP address as the URL parameter)",
		RunE:  runVhost,
	}
	if err := addCommonHTTPOptions(cmdVhost); err != nil {
		log.Fatalf("%v", err)
	}
	cmdVhost.Flags().StringP("status-codes", "s", "", "Positive status codes (will be overwritten with status-codes-blacklist if set). Can also handle ranges like 200,300-400,404.")
	cmdVhost.Flags().StringP("status-codes-blacklist", "b", "404", "Negative status codes (will override status-codes if set). Can also handle ranges like 200,300-400,404.")
	cmdVhost.Flags().BoolP("append-domain", "", false, "Append main domain from URL to words from wordlist. Otherwise the fully qualified domains need to be specified in the wordlist.")
	cmdVhost.Flags().String("exclude-length", "", "exclude the following content lengths (completely ignores the status). You can separate multiple lengths by comma and it also supports ranges like 203-206")
	cmdVhost.Flags().String("domain", "", "the domain to append when using an IP address as URL. If left empty and you specify a domain based URL the hostname from the URL is extracted")

	cmdVhost.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdVhost)
}
