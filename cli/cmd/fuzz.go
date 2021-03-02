package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterfuzz"
	"github.com/OJ/gobuster/v3/helper"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

// nolint:gochecknoglobals
var cmdFuzz *cobra.Command

func runFuzz(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseFuzzOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %w", err)
	}

	plugin, err := gobusterfuzz.NewGobusterFuzz(globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobusterfuzz: %w", err)
	}

	if err := cli.Gobuster(mainContext, globalopts, plugin); err != nil {
		var wErr *gobusterfuzz.ErrWildcard
		if errors.As(err, &wErr) {
			return fmt.Errorf("%w. To continue please exclude the status code or the length", wErr)
		}
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}

func parseFuzzOptions() (*libgobuster.Options, *gobusterfuzz.OptionsFuzz, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}

	plugin := gobusterfuzz.NewOptionsFuzz()

	httpOpts, err := parseCommonHTTPOptions(cmdFuzz)
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

	// blacklist will override the normal status codes
	plugin.ExcludedStatusCodes, err = cmdFuzz.Flags().GetString("excludestatuscodes")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for excludestatuscodes: %w", err)
	}
	ret, err := helper.ParseCommaSeparatedInt(plugin.ExcludedStatusCodes)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for excludestatuscodes: %w", err)
	}
	plugin.ExcludedStatusCodesParsed = ret

	plugin.ExcludeLength, err = cmdFuzz.Flags().GetIntSlice("exclude-length")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for excludelength: %w", err)
	}

	return globalopts, plugin, nil
}

// nolint:gochecknoinits
func init() {
	cmdFuzz = &cobra.Command{
		Use:   "fuzz",
		Short: "Uses fuzzing mode",
		RunE:  runFuzz,
	}

	if err := addCommonHTTPOptions(cmdFuzz); err != nil {
		log.Fatalf("%v", err)
	}
	cmdFuzz.Flags().StringP("excludestatuscodes", "b", "", "Negative status codes (will override statuscodes if set)")
	cmdFuzz.Flags().IntSlice("exclude-length", []int{}, "exclude the following content length (completely ignores the status). Supply multiple times to exclude multiple sizes.")

	cmdFuzz.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdFuzz)
}
