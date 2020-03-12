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
		return fmt.Errorf("error on parsing arguments: %v", err)
	}

	plugin, err := gobusterfuzz.NewGobusterFuzz(mainContext, globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobusterfuzz: %v", err)
	}

	if err := cli.Gobuster(mainContext, globalopts, plugin); err != nil {
		if goberr, ok := err.(*gobusterfuzz.ErrWildcard); ok {
			return fmt.Errorf("%s. To force processing of Wildcard responses, specify the '--wildcard' switch", goberr.Error())
		}
		return fmt.Errorf("error on running gobuster: %v", err)
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

	plugin.Extensions, err = cmdFuzz.Flags().GetString("extensions")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for extensions: %v", err)
	}

	if plugin.Extensions != "" {
		ret, err := helper.ParseExtensions(plugin.Extensions)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for extensions: %v", err)
		}
		plugin.ExtensionsParsed = ret
	}

	plugin.StatusCodesBlacklist, err = cmdFuzz.Flags().GetString("statuscodesblacklist")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for statuscodesblacklist: %v", err)
	}

	// blacklist will override the normal status codes
	if plugin.StatusCodesBlacklist != "" {
		ret, err := helper.ParseCommaSeperatedInt(plugin.StatusCodesBlacklist)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for statuscodesblacklist: %v", err)
		}
		plugin.StatusCodesBlacklistParsed = ret
	} else {
		// parse normal status codes
		plugin.StatusCodes, err = cmdFuzz.Flags().GetString("statuscodes")
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for statuscodes: %v", err)
		}
		ret, err := helper.ParseCommaSeperatedInt(plugin.StatusCodes)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for statuscodes: %v", err)
		}
		plugin.StatusCodesParsed = ret
	}

	plugin.UseSlash, err = cmdFuzz.Flags().GetBool("addslash")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for addslash: %v", err)
	}

	plugin.Expanded, err = cmdFuzz.Flags().GetBool("expanded")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for expanded: %v", err)
	}

	plugin.NoStatus, err = cmdFuzz.Flags().GetBool("nostatus")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for nostatus: %v", err)
	}

	plugin.IncludeLength, err = cmdFuzz.Flags().GetBool("includelength")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for includelength: %v", err)
	}

	plugin.WildcardForced, err = cmdFuzz.Flags().GetBool("wildcard")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for wildcard: %v", err)
	}

	plugin.DiscoverBackup, err = cmdFuzz.Flags().GetBool("discoverbackup")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for discoverbackup: %v", err)
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
	cmdFuzz.Flags().StringP("statuscodes", "s", "200,204,301,302,307,401,403", "Positive status codes (will be overwritten with statuscodesblacklist if set)")
	cmdFuzz.Flags().StringP("statuscodesblacklist", "b", "", "Negative status codes (will override statuscodes if set)")
	cmdFuzz.Flags().StringP("extensions", "x", "", "File extension(s) to search for")
	cmdFuzz.Flags().BoolP("expanded", "e", false, "Expanded mode, print full URLs")
	cmdFuzz.Flags().BoolP("nostatus", "n", false, "Don't print status codes")
	cmdFuzz.Flags().BoolP("includelength", "l", false, "Include the length of the body in the output")
	cmdFuzz.Flags().BoolP("addslash", "f", false, "Append / to each request")
	cmdFuzz.Flags().BoolP("wildcard", "", false, "Force continued operation when wildcard found")
	cmdFuzz.Flags().BoolP("discoverbackup", "d", false, "Upon finding a file search for backup files")

	cmdFuzz.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdFuzz)
}
