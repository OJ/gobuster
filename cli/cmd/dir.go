package cmd

import (
	"fmt"
	"log"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterdir"
	"github.com/OJ/gobuster/v3/helper"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

var cmdDir *cobra.Command

func runDir(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseDirOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %w", err)
	}

	plugin, err := gobusterdir.NewGobusterDir(mainContext, globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobusterdir: %w", err)
	}

	if err := cli.Gobuster(mainContext, globalopts, plugin); err != nil {
		if goberr, ok := err.(*gobusterdir.ErrWildcard); ok {
			return fmt.Errorf("%s. To force processing of Wildcard responses, specify the '--wildcard' switch", goberr.Error())
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
	plugin.InsecureSSL = httpOpts.InsecureSSL
	plugin.Headers = httpOpts.Headers
	plugin.Method = httpOpts.Method

	plugin.Extensions, err = cmdDir.Flags().GetString("extensions")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for extensions: %w", err)
	}

	if plugin.Extensions != "" {
		ret, err := helper.ParseExtensions(plugin.Extensions)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for extensions: %w", err)
		}
		plugin.ExtensionsParsed = ret
	}

	plugin.StatusCodesBlacklist, err = cmdDir.Flags().GetString("statuscodesblacklist")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for statuscodesblacklist: %w", err)
	}

	// blacklist will override the normal status codes
	if plugin.StatusCodesBlacklist != "" {
		ret, err := helper.ParseCommaSeperatedInt(plugin.StatusCodesBlacklist)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for statuscodesblacklist: %w", err)
		}
		plugin.StatusCodesBlacklistParsed = ret
	} else {
		// parse normal status codes
		plugin.StatusCodes, err = cmdDir.Flags().GetString("statuscodes")
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for statuscodes: %w", err)
		}
		ret, err := helper.ParseCommaSeperatedInt(plugin.StatusCodes)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for statuscodes: %w", err)
		}
		plugin.StatusCodesParsed = ret
	}

	plugin.UseSlash, err = cmdDir.Flags().GetBool("addslash")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for addslash: %w", err)
	}

	plugin.Expanded, err = cmdDir.Flags().GetBool("expanded")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for expanded: %w", err)
	}

	plugin.NoStatus, err = cmdDir.Flags().GetBool("nostatus")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for nostatus: %w", err)
	}

	plugin.IncludeLength, err = cmdDir.Flags().GetBool("includelength")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for includelength: %w", err)
	}

	plugin.WildcardForced, err = cmdDir.Flags().GetBool("wildcard")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for wildcard: %w", err)
	}

	plugin.DiscoverBackup, err = cmdDir.Flags().GetBool("discoverbackup")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for discoverbackup: %w", err)
	}

	return globalopts, plugin, nil
}

func init() {
	cmdDir = &cobra.Command{
		Use:   "dir",
		Short: "Uses directory/file enumeration mode",
		RunE:  runDir,
	}

	if err := addCommonHTTPOptions(cmdDir); err != nil {
		log.Fatalf("%v", err)
	}
	cmdDir.Flags().StringP("statuscodes", "s", "200,204,301,302,307,401,403", "Positive status codes (will be overwritten with statuscodesblacklist if set)")
	cmdDir.Flags().StringP("statuscodesblacklist", "b", "", "Negative status codes (will override statuscodes if set)")
	cmdDir.Flags().StringP("extensions", "x", "", "File extension(s) to search for")
	cmdDir.Flags().BoolP("expanded", "e", false, "Expanded mode, print full URLs")
	cmdDir.Flags().BoolP("nostatus", "n", false, "Don't print status codes")
	cmdDir.Flags().BoolP("includelength", "l", false, "Include the length of the body in the output")
	cmdDir.Flags().BoolP("addslash", "f", false, "Append / to each request")
	cmdDir.Flags().BoolP("wildcard", "", false, "Force continued operation when wildcard found")
	cmdDir.Flags().BoolP("discoverbackup", "d", false, "Upon finding a file search for backup files")

	cmdDir.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdDir)
}
