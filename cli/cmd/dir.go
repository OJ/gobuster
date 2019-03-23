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
	"github.com/OJ/gobuster/v3/gobusterdir"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var cmdDir *cobra.Command


func runDir(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseDirOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	plugin, err := gobusterdir.NewGobusterDir(ctx, globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("error on creating gobusterdir: %v", err)
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

func parseDirOptions() (*libgobuster.Options, *gobusterdir.OptionsDir, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}

	plugin := gobusterdir.NewOptionsDir()

	plugin.URL, err = cmdDir.Flags().GetString("url")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for url: %v", err)
	}

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

	plugin.StatusCodes, err = cmdDir.Flags().GetString("statuscodes")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for statuscodes: %v", err)
	}

	if err = plugin.ParseStatusCodes(); err != nil {
		return nil, nil, fmt.Errorf("invalid value for statuscodes: %v", err)
	}

	plugin.Cookies, err = cmdDir.Flags().GetString("cookies")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for cookies: %v", err)
	}

	plugin.Username, err = cmdDir.Flags().GetString("username")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for username: %v", err)
	}

	plugin.Password, err = cmdDir.Flags().GetString("password")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for password: %v", err)
	}

	plugin.Extensions, err = cmdDir.Flags().GetString("extensions")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for extensions: %v", err)
	}

	if plugin.Extensions != "" {
		if err = plugin.ParseExtensions(); err != nil {
			return nil, nil, fmt.Errorf("invalid value for extensions: %v", err)
		}
	}

	plugin.UserAgent, err = cmdDir.Flags().GetString("useragent")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for useragent: %v", err)
	}

	plugin.Proxy, err = cmdDir.Flags().GetString("proxy")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for proxy: %v", err)
	}

	plugin.Timeout, err = cmdDir.Flags().GetDuration("timeout")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for timeout: %v", err)
	}

	plugin.FollowRedirect, err = cmdDir.Flags().GetBool("followredirect")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for followredirect: %v", err)
	}

	plugin.Expanded, err = cmdDir.Flags().GetBool("expanded")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for expanded: %v", err)
	}

	plugin.NoStatus, err = cmdDir.Flags().GetBool("nostatus")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for nostatus: %v", err)
	}

	plugin.IncludeLength, err = cmdDir.Flags().GetBool("includelength")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for includelength: %v", err)
	}

	plugin.InsecureSSL, err = cmdDir.Flags().GetBool("insecuressl")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for insecuressl: %v", err)
	}

	plugin.WildcardForced, err = cmdDir.Flags().GetBool("wildcard")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for wildcard: %v", err)
	}

	plugin.UseSlash, err = cmdDir.Flags().GetBool("addslash")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for addslash: %v", err)
	}

	plugin.Headers, err = cmdDir.Flags().GetStringArray("headers")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for headers: %v", err)
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

	return globalopts, plugin, nil
}

func init() {
	cmdDir = &cobra.Command{
		Use:   "dir",
		Short: "Uses directory/file brutceforcing mode",
		RunE:  runDir,
	}

	cmdDir.Flags().StringP("url", "u", "", "The target URL")
	cmdDir.Flags().StringP("statuscodes", "s", "200,204,301,302,307,401,403", "Positive status codes")
	cmdDir.Flags().StringP("cookies", "c", "", "Cookies to use for the requests")
	cmdDir.Flags().StringP("username", "U", "", "Username for Basic Auth")
	cmdDir.Flags().StringP("password", "P", "", "Password for Basic Auth")
	cmdDir.Flags().StringP("extensions", "x", "", "File extension(s) to search for")
	cmdDir.Flags().StringP("useragent", "a", libgobuster.DefaultUserAgent(), "Set the User-Agent string")
	cmdDir.Flags().StringP("proxy", "p", "", "Proxy to use for requests [http(s)://host:port]")
	cmdDir.Flags().DurationP("timeout", "", 10*time.Second, "HTTP Timeout")
	cmdDir.Flags().BoolP("followredirect", "r", false, "Follow redirects")
	cmdDir.Flags().BoolP("expanded", "e", false, "Expanded mode, print full URLs")
	cmdDir.Flags().BoolP("nostatus", "n", false, "Don't print status codes")
	cmdDir.Flags().BoolP("includelength", "l", false, "Include the length of the body in the output")
	cmdDir.Flags().BoolP("insecuressl", "k", false, "Skip SSL certificate verification")
	cmdDir.Flags().BoolP("addslash", "f", false, "Apped / to each request")
	cmdDir.Flags().BoolP("wildcard", "", false, "Force continued operation when wildcard found")
	cmdDir.Flags().StringArrayP("headers","H",[]string{""},"Specify HTTP headers, -H 'Header1: val1' -H 'Header2: val2'")

	if err := cmdDir.MarkFlagRequired("url"); err != nil {
		log.Fatalf("error on marking flag as required: %v", err)
	}

	cmdDir.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdDir)
}
