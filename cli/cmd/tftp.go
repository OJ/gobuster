package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobustertftp"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

// nolint:gochecknoglobals
var cmdTFTP *cobra.Command

func runTFTP(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseTFTPOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %w", err)
	}

	plugin, err := gobustertftp.NewGobusterTFTP(globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobustertftp: %w", err)
	}

	log := libgobuster.NewLogger(globalopts.Debug)
	if err := cli.Gobuster(mainContext, globalopts, plugin, log); err != nil {
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}

func parseTFTPOptions() (*libgobuster.Options, *gobustertftp.OptionsTFTP, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}
	pluginOpts := gobustertftp.NewOptionsTFTP()

	pluginOpts.Server, err = cmdTFTP.Flags().GetString("server")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for domain: %w", err)
	}

	if !strings.Contains(pluginOpts.Server, ":") {
		pluginOpts.Server = fmt.Sprintf("%s:69", pluginOpts.Server)
	}

	pluginOpts.Timeout, err = cmdTFTP.Flags().GetDuration("timeout")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for timeout: %w", err)
	}

	return globalopts, pluginOpts, nil
}

// nolint:gochecknoinits
func init() {
	cmdTFTP = &cobra.Command{
		Use:   "tftp",
		Short: "Uses TFTP enumeration mode",
		RunE:  runTFTP,
	}

	cmdTFTP.Flags().StringP("server", "s", "", "The target TFTP server")
	cmdTFTP.Flags().DurationP("timeout", "", time.Second, "TFTP timeout")
	if err := cmdTFTP.MarkFlagRequired("server"); err != nil {
		log.Fatalf("error on marking flag as required: %v", err)
	}

	cmdTFTP.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdTFTP)
}
