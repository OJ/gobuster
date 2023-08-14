package cmd

import (
	"fmt"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobustergcs"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

// nolint:gochecknoglobals
var cmdGCS *cobra.Command

func runGCS(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseGCSOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %w", err)
	}

	plugin, err := gobustergcs.NewGobusterGCS(globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobustergcs: %w", err)
	}

	log := libgobuster.NewLogger(globalopts.Debug)
	if err := cli.Gobuster(mainContext, globalopts, plugin, log); err != nil {
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}

func parseGCSOptions() (*libgobuster.Options, *gobustergcs.OptionsGCS, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}

	pluginopts := gobustergcs.NewOptionsGCS()

	httpOpts, err := parseBasicHTTPOptions(cmdGCS)
	if err != nil {
		return nil, nil, err
	}

	pluginopts.UserAgent = httpOpts.UserAgent
	pluginopts.Proxy = httpOpts.Proxy
	pluginopts.Timeout = httpOpts.Timeout
	pluginopts.NoTLSValidation = httpOpts.NoTLSValidation
	pluginopts.RetryOnTimeout = httpOpts.RetryOnTimeout
	pluginopts.RetryAttempts = httpOpts.RetryAttempts
	pluginopts.TLSCertificate = httpOpts.TLSCertificate

	pluginopts.MaxFilesToList, err = cmdGCS.Flags().GetInt("maxfiles")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for maxfiles: %w", err)
	}

	return globalopts, pluginopts, nil
}

// nolint:gochecknoinits
func init() {
	cmdGCS = &cobra.Command{
		Use:   "gcs",
		Short: "Uses gcs bucket enumeration mode",
		RunE:  runGCS,
	}

	addBasicHTTPOptions(cmdGCS)
	cmdGCS.Flags().IntP("maxfiles", "m", 5, "max files to list when listing buckets (only shown in verbose mode)")

	cmdGCS.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdGCS)
}
