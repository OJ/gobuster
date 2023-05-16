package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterdir"
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

	log := libgobuster.NewLogger(globalopts.Debug)
	if err := cli.Gobuster(mainContext, globalopts, plugin, log); err != nil {
		var wErr *gobusterdir.ErrWildcard
		if errors.As(err, &wErr) {
			return fmt.Errorf("%w. To continue please exclude the status code or the length", wErr)
		}
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}

func parseDirOptions() (*libgobuster.Options, *gobusterdir.OptionsDir, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}

	pluginOpts := gobusterdir.NewOptionsDir()

	httpOpts, err := parseCommonHTTPOptions(cmdDir)
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

	pluginOpts.Extensions, err = cmdDir.Flags().GetString("extensions")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for extensions: %w", err)
	}

	ret, err := libgobuster.ParseExtensions(pluginOpts.Extensions)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for extensions: %w", err)
	}
	pluginOpts.ExtensionsParsed = ret

	pluginOpts.ExtensionsFile, err = cmdDir.Flags().GetString("extensions-file")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for extensions file: %w", err)
	}

	if pluginOpts.ExtensionsFile != "" {
		extensions, err := libgobuster.ParseExtensionsFile(pluginOpts.ExtensionsFile)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for extensions file: %w", err)
		}
		pluginOpts.ExtensionsParsed.AddRange(extensions)
	}

	// parse normal status codes
	pluginOpts.StatusCodes, err = cmdDir.Flags().GetString("status-codes")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for status-codes: %w", err)
	}
	ret2, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.StatusCodes)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for status-codes: %w", err)
	}
	pluginOpts.StatusCodesParsed = ret2

	// blacklist will override the normal status codes
	pluginOpts.StatusCodesBlacklist, err = cmdDir.Flags().GetString("status-codes-blacklist")
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

	pluginOpts.UseSlash, err = cmdDir.Flags().GetBool("add-slash")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for add-slash: %w", err)
	}

	pluginOpts.Expanded, err = cmdDir.Flags().GetBool("expanded")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for expanded: %w", err)
	}

	pluginOpts.NoStatus, err = cmdDir.Flags().GetBool("no-status")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for no-status: %w", err)
	}

	pluginOpts.HideLength, err = cmdDir.Flags().GetBool("hide-length")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for hide-length: %w", err)
	}

	pluginOpts.DiscoverBackup, err = cmdDir.Flags().GetBool("discover-backup")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for discover-backup: %w", err)
	}

	pluginOpts.ExcludeLength, err = cmdDir.Flags().GetString("exclude-length")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for exclude-length: %w", err)
	}
	ret4, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.ExcludeLength)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for exclude-length: %w", err)
	}
	pluginOpts.ExcludeLengthParsed = ret4

	return globalopts, pluginOpts, nil
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
	cmdDir.Flags().StringP("status-codes", "s", "", "Positive status codes (will be overwritten with status-codes-blacklist if set). Can also handle ranges like 200,300-400,404.")
	cmdDir.Flags().StringP("status-codes-blacklist", "b", "404", "Negative status codes (will override status-codes if set). Can also handle ranges like 200,300-400,404.")
	cmdDir.Flags().StringP("extensions", "x", "", "File extension(s) to search for")
	cmdDir.Flags().StringP("extensions-file", "X", "", "Read file extension(s) to search from the file")
	cmdDir.Flags().BoolP("expanded", "e", false, "Expanded mode, print full URLs")
	cmdDir.Flags().BoolP("no-status", "n", false, "Don't print status codes")
	cmdDir.Flags().Bool("hide-length", false, "Hide the length of the body in the output")
	cmdDir.Flags().BoolP("add-slash", "f", false, "Append / to each request")
	cmdDir.Flags().BoolP("discover-backup", "d", false, "Also search for backup files by appending multiple backup extensions")
	cmdDir.Flags().String("exclude-length", "", "exclude the following content lengths (completely ignores the status). You can separate multiple lengths by comma and it also supports ranges like 203-206")

	cmdDir.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdDir)
}
