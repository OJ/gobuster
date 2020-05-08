package cmd

import (
	"fmt"
	"log"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterfuzz"
	"github.com/OJ/gobuster/v3/helper"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

var cmdFuzz *cobra.Command

func runFuzz(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseFuzzOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %w", err)
	}

	plugin, err := gobusterfuzz.NewGobusterFuzz(mainContext, globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobusterfuzz: %w", err)
	}

	if err := cli.Gobuster(mainContext, globalopts, plugin); err != nil {
		if goberr, ok := err.(*gobusterfuzz.ErrWildcard); ok {
			return fmt.Errorf("%s. To force processing of Wildcard responses, specify the '--wildcard' switch", goberr.Error())
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
	plugin.InsecureSSL = httpOpts.InsecureSSL
	plugin.Headers = httpOpts.Headers
	plugin.Method = httpOpts.Method

	plugin.ExcludedStatusCodes, err = cmdFuzz.Flags().GetString("excludestatuscodes")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for excludestatuscodes: %w", err)
	}

	// blacklist will override the normal status codes
	if plugin.ExcludedStatusCodes != "" {
		ret, err := helper.ParseCommaSeperatedInt(plugin.ExcludedStatusCodes)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for excludestatuscodes: %w", err)
		}
		plugin.ExcludedStatusCodesParsed = ret
	}

	plugin.WildcardForced, err = cmdFuzz.Flags().GetBool("wildcard")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for wildcard: %w", err)
	}

	plugin.ExcludeLength, err = cmdFuzz.Flags().GetIntSlice("exclude-length")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for excludelength: %w", err)
	}

	return globalopts, plugin, nil
}

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
	cmdFuzz.Flags().BoolP("wildcard", "", false, "Force continued operation when wildcard found")

	cmdFuzz.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdFuzz)
}
