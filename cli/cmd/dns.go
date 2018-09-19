package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/OJ/gobuster/cli"
	"github.com/OJ/gobuster/gobusterdns"
	"github.com/OJ/gobuster/libgobuster"
	"github.com/spf13/cobra"
)

var cmdDNS *cobra.Command

func runDNS(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseDNS()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	plugin, err := gobusterdns.NewGobusterDNS(globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("Error on creating gobusterdns: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			// caught CTRL+C
			if !globalopts.Quiet {
				fmt.Println("\n[!] Keyboard interrupt detected, terminating.")
			}
			cancel()
		}
	}()

	if err := cli.Gobuster(ctx, globalopts, plugin); err != nil {
		return fmt.Errorf("error on running goubster: %v", err)
	}
	return nil
}

func parseDNS() (*libgobuster.Options, *gobusterdns.OptionsDNS, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}
	plugin := gobusterdns.NewOptionsDNS()

	domain, err := cmdDNS.Flags().GetString("domain")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for domain: %v", err)
	}
	plugin.Domain = domain

	showips, err := cmdDNS.Flags().GetBool("showips")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for showips: %v", err)
	}
	plugin.ShowIPs = showips

	showcname, err := cmdDNS.Flags().GetBool("showcname")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for showcname: %v", err)
	}
	plugin.ShowCNAME = showcname

	wildcard, err := cmdDNS.Flags().GetBool("wildcard")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for wildcard: %v", err)
	}
	plugin.WildcardForced = wildcard

	return globalopts, plugin, nil
}

func init() {
	cmdDNS = &cobra.Command{
		Use:   "dns",
		Short: "uses dns mode",
		RunE:  runDNS,
	}

	cmdDNS.Flags().StringP("domain", "d", "", "The target domain")
	cmdDNS.Flags().BoolP("showips", "i", false, "Show IP addresses")
	cmdDNS.Flags().BoolP("showcname", "c", false, "Show CNAME records (cannot be used with '-i' option)")
	cmdDNS.Flags().BoolP("wildcard", "", false, "Force continued operation when wildcard found")
	if err := cmdDNS.MarkFlagRequired("domain"); err != nil {
		log.Fatalf("error on marking flag as required: %v", err)
	}
	rootCmd.AddCommand(cmdDNS)
}
