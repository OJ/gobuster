package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobustervhost"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var cmdVhost *cobra.Command

func runVhost(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseVhostOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	plugin, err := gobustervhost.NewGobusterVhost(ctx, globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobustervhost: %v", err)
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

func parseVhostOptions() (*libgobuster.Options, *gobustervhost.OptionsVhost, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}
	var plugin gobustervhost.OptionsVhost

	url, err := cmdVhost.Flags().GetString("url")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for url: %v", err)
	}
	plugin.URL = url
	if !strings.HasPrefix(plugin.URL, "http") {
		// check to see if a port was specified
		re := regexp.MustCompile(`^[^/]+:(\d+)`)
		match := re.FindStringSubmatch(plugin.URL)

		if len(match) < 2 {
			// no port, default to http on 80
			plugin.URL = fmt.Sprintf("http://%s", plugin.URL)
		} else {
			port, err2 := strconv.Atoi(match[1])
			if err2 != nil || (port != 80 && port != 443) {
				return nil, nil, fmt.Errorf("url scheme not specified")
			} else if port == 80 {
				plugin.URL = fmt.Sprintf("http://%s", plugin.URL)
			} else {
				plugin.URL = fmt.Sprintf("https://%s", plugin.URL)
			}
		}
	}

	plugin.Cookies, err = cmdVhost.Flags().GetString("cookies")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for cookies: %v", err)
	}

	plugin.Username, err = cmdVhost.Flags().GetString("username")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for username: %v", err)
	}

	plugin.Password, err = cmdVhost.Flags().GetString("password")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for password: %v", err)
	}

	plugin.UserAgent, err = cmdVhost.Flags().GetString("useragent")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for useragent: %v", err)
	}

	plugin.Proxy, err = cmdVhost.Flags().GetString("proxy")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for proxy: %v", err)
	}

	plugin.Timeout, err = cmdVhost.Flags().GetDuration("timeout")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for timeout: %v", err)
	}

	plugin.FollowRedirect, err = cmdDir.Flags().GetBool("followredirect")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for followredirect: %v", err)
	}

	plugin.InsecureSSL, err = cmdVhost.Flags().GetBool("insecuressl")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for insecuressl: %v", err)
	}

	// Prompt for PW if not provided
	if plugin.Username != "" && plugin.Password == "" {
		fmt.Printf("[?] Auth Password: ")
		passBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		// print a newline to simulate the newline that was entered
		// this means that formatting/printing after doesn't look bad.
		fmt.Println("")
		if err != nil {
			return nil, nil, fmt.Errorf("username given but reading of password failed")
		}
		plugin.Password = string(passBytes)
	}
	// if it's still empty bail out
	if plugin.Username != "" && plugin.Password == "" {
		return nil, nil, fmt.Errorf("username was provided but password is missing")
	}

	return globalopts, &plugin, nil
}

func init() {
	cmdVhost = &cobra.Command{
		Use:   "vhost",
		Short: "Uses VHOST bruteforcing mode",
		RunE:  runVhost,
	}
	cmdVhost.Flags().StringP("url", "u", "", "The target URL")
	cmdVhost.Flags().StringP("cookies", "c", "", "Cookies to use for the requests")
	cmdVhost.Flags().StringP("username", "U", "", "Username for Basic Auth")
	cmdVhost.Flags().StringP("password", "P", "", "Password for Basic Auth")
	cmdVhost.Flags().StringP("useragent", "a", libgobuster.DefaultUserAgent(), "Set the User-Agent string")
	cmdVhost.Flags().StringP("proxy", "p", "", "Proxy to use for requests [http(s)://host:port]")
	cmdVhost.Flags().DurationP("timeout", "", 10*time.Second, "HTTP Timeout")
	cmdVhost.Flags().BoolP("followredirect", "r", true, "Follow redirects")
	cmdVhost.Flags().BoolP("insecuressl", "k", false, "Skip SSL certificate verification")
	if err := cmdVhost.MarkFlagRequired("url"); err != nil {
		log.Fatalf("error on marking flag as required: %v", err)
	}

	cmdVhost.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdVhost)
}
