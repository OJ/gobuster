package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

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

	if !containsFuzzKeyword(*pluginopts) {
		return fmt.Errorf("please provide the %s keyword", gobusterfuzz.FuzzKeyword)
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

	pluginOpts := gobusterfuzz.NewOptionsFuzz()

	httpOpts, err := parseCommonHTTPOptions(cmdFuzz)
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

	// blacklist will override the normal status codes
	pluginOpts.ExcludedStatusCodes, err = cmdFuzz.Flags().GetString("excludestatuscodes")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for excludestatuscodes: %w", err)
	}
	ret, err := helper.ParseCommaSeparatedInt(pluginOpts.ExcludedStatusCodes)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for excludestatuscodes: %w", err)
	}
	pluginOpts.ExcludedStatusCodesParsed = ret

	pluginOpts.ExcludeLength, err = cmdFuzz.Flags().GetIntSlice("exclude-length")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for excludelength: %w", err)
	}

	pluginOpts.RequestBody, err = cmdFuzz.Flags().GetString("body")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for body: %w", err)
	}

	return globalopts, pluginOpts, nil
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
	cmdFuzz.Flags().StringP("body", "B", "", "Request body")

	cmdFuzz.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdFuzz)
}

func containsFuzzKeyword(pluginopts gobusterfuzz.OptionsFuzz) bool {
	if strings.Contains(pluginopts.URL, gobusterfuzz.FuzzKeyword) {
		return true
	}

	if strings.Contains(pluginopts.RequestBody, gobusterfuzz.FuzzKeyword) {
		return true
	}

	for _, h := range pluginopts.Headers {
		if strings.Contains(h.Name, gobusterfuzz.FuzzKeyword) || strings.Contains(h.Value, gobusterfuzz.FuzzKeyword) {
			return true
		}
	}

	return false
}
