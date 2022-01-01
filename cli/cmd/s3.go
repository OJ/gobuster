package cmd

import (
	"fmt"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/lib"
	"github.com/OJ/gobuster/v3/s3"
	"github.com/spf13/cobra"
)

// nolint:gochecknoglobals
var cmdS3 *cobra.Command

func runS3(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseS3Options()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %w", err)
	}

	plugin, err := s3.NewGobusterS3(globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobusters3: %w", err)
	}

	if err := cli.Gobuster(mainContext, globalopts, plugin); err != nil {
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}

func parseS3Options() (*lib.Options, *s3.OptionsS3, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}

	plugin := s3.NewOptionsS3()

	httpOpts, err := parseBasicHTTPOptions(cmdS3)
	if err != nil {
		return nil, nil, err
	}

	plugin.UserAgent = httpOpts.UserAgent
	plugin.Proxy = httpOpts.Proxy
	plugin.Timeout = httpOpts.Timeout
	plugin.NoTLSValidation = httpOpts.NoTLSValidation
	plugin.RetryOnTimeout = httpOpts.RetryOnTimeout
	plugin.RetryAttempts = httpOpts.RetryAttempts

	plugin.MaxFilesToList, err = cmdS3.Flags().GetInt("maxfiles")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for maxfiles: %w", err)
	}

	return globalopts, plugin, nil
}

// nolint:gochecknoinits
func init() {
	cmdS3 = &cobra.Command{
		Use:   "s3",
		Short: "Uses aws bucket enumeration mode",
		RunE:  runS3,
	}

	addBasicHTTPOptions(cmdS3)
	cmdS3.Flags().IntP("maxfiles", "m", 5, "max files to list when listing buckets (only shown in verbose mode)")

	cmdS3.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdS3)
}
