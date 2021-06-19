package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterdir"
	"github.com/OJ/gobuster/v3/helper"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

// nolint:gochecknoglobals
var cmdDir *cobra.Command

func runDir(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseDirOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %w", err)
	}

	plugin, err := gobusterdir.NewGobusterDir(globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobusterdir: %w", err)
	}

	if err := cli.Gobuster(mainContext, globalopts, plugin); err != nil {
		var wErr *gobusterdir.ErrWildcard
		if errors.As(err, &wErr) {
			return fmt.Errorf("%w. To continue please exclude the status code or the length", wErr)
		}
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}

func parseDirOptions() (*libgobuster.Options, *gobusterdir.OptionsDir, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}

	plugin := gobusterdir.NewOptionsDir()

	httpOpts, err := parseCommonHTTPOptions(cmdDir)
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

	plugin.Extensions, err = cmdDir.Flags().GetString("extensions")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for extensions: %w", err)
	}
	ret, err := helper.ParseExtensions(plugin.Extensions)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for extensions: %w", err)
	}
	plugin.ExtensionsParsed = ret

	// parse normal status codes
	plugin.StatusCodes, err = cmdDir.Flags().GetString("status-codes")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for status-codes: %w", err)
	}
	ret2, err := helper.ParseCommaSeparatedInt(plugin.StatusCodes)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for status-codes: %w", err)
	}
	plugin.StatusCodesParsed = ret2

	// blacklist will override the normal status codes
	plugin.StatusCodesBlacklist, err = cmdDir.Flags().GetString("status-codes-blacklist")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for status-codes-blacklist: %w", err)
	}
	ret3, err := helper.ParseCommaSeparatedInt(plugin.StatusCodesBlacklist)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for status-codes-blacklist: %w", err)
	}
	plugin.StatusCodesBlacklistParsed = ret3

	if plugin.StatusCodes != "" && plugin.StatusCodesBlacklist != "" {
		return nil, nil, fmt.Errorf("status-codes and status-codes-blacklist are both set, please set only one")
	}

	if plugin.StatusCodes == "" && plugin.StatusCodesBlacklist == "" {
		return nil, nil, fmt.Errorf("status-codes and status-codes-blacklist are both not set, please set one")
	}

	plugin.UseSlash, err = cmdDir.Flags().GetBool("add-slash")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for add-slash: %w", err)
	}

	plugin.Expanded, err = cmdDir.Flags().GetBool("expanded")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for expanded: %w", err)
	}

	plugin.NoStatus, err = cmdDir.Flags().GetBool("no-status")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for no-status: %w", err)
	}

	plugin.HideLength, err = cmdDir.Flags().GetBool("hide-length")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for hide-length: %w", err)
	}

	plugin.DiscoverBackup, err = cmdDir.Flags().GetBool("discover-backup")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for discover-backup: %w", err)
	}

	plugin.ExcludeLength, err = cmdDir.Flags().GetIntSlice("exclude-length")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for excludelength: %w", err)
	}

	return globalopts, plugin, nil
}

// nolint:gochecknoinits
func init() {
	cmdDir = &cobra.Command{
		Use:   "dir",
		Short: "Uses directory/file enumeration mode",
		RunE:  runDir,
	}

	if err := addCommonHTTPOptions(cmdDir); err != nil {
		log.Fatalf("%v", err)
	}
	cmdDir.Flags().StringP("status-codes", "s", "", "Positive status codes (will be overwritten with status-codes-blacklist if set)")
	cmdDir.Flags().StringP("status-codes-blacklist", "b", "404", "Negative status codes (will override status-codes if set)")
	cmdDir.Flags().StringP("extensions", "x", "", "File extension(s) to search for")
	cmdDir.Flags().BoolP("expanded", "e", false, "Expanded mode, print full URLs")
	cmdDir.Flags().BoolP("no-status", "n", false, "Don't print status codes")
	cmdDir.Flags().Bool("hide-length", false, "Hide the length of the body in the output")
	cmdDir.Flags().BoolP("add-slash", "f", false, "Append / to each request")
	cmdDir.Flags().BoolP("discover-backup", "d", false, "Upon finding a file search for backup files")
	cmdDir.Flags().IntSlice("exclude-length", []int{}, "exclude the following content length (completely ignores the status). Supply multiple times to exclude multiple sizes.")

	cmdDir.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdDir)
}
