package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func addCommonHTTPOptions(cmd *cobra.Command) error {
	cmd.Flags().StringP("url", "u", "", "The target URL")
	cmd.Flags().StringP("cookies", "c", "", "Cookies to use for the requests")
	cmd.Flags().StringP("username", "U", "", "Username for Basic Auth")
	cmd.Flags().StringP("password", "P", "", "Password for Basic Auth")
	cmd.Flags().StringP("useragent", "a", libgobuster.DefaultUserAgent(), "Set the User-Agent string")
	cmd.Flags().StringP("proxy", "p", "", "Proxy to use for requests [http(s)://host:port]")
	cmd.Flags().DurationP("timeout", "", 10*time.Second, "HTTP Timeout")
	cmd.Flags().BoolP("followredirect", "r", false, "Follow redirects")
	cmd.Flags().BoolP("insecuressl", "k", false, "Skip SSL certificate verification")
	cmd.Flags().StringArrayP("headers", "H", []string{""}, "Specify HTTP headers, -H 'Header1: val1' -H 'Header2: val2'")

	if err := cmdDir.MarkFlagRequired("url"); err != nil {
		return fmt.Errorf("error on marking flag as required: %v", err)
	}

	return nil
}

func parseCommonHTTPOptions(cmd *cobra.Command) (libgobuster.OptionsHTTP, error) {
	options := libgobuster.OptionsHTTP{}
	var err error

	options.URL, err = cmd.Flags().GetString("url")
	if err != nil {
		return options, fmt.Errorf("invalid value for url: %v", err)
	}

	if !strings.HasPrefix(options.URL, "http") {
		// check to see if a port was specified
		re := regexp.MustCompile(`^[^/]+:(\d+)`)
		match := re.FindStringSubmatch(options.URL)

		if len(match) < 2 {
			// no port, default to http on 80
			options.URL = fmt.Sprintf("http://%s", options.URL)
		} else {
			port, err2 := strconv.Atoi(match[1])
			if err2 != nil || (port != 80 && port != 443) {
				return options, fmt.Errorf("url scheme not specified")
			} else if port == 80 {
				options.URL = fmt.Sprintf("http://%s", options.URL)
			} else {
				options.URL = fmt.Sprintf("https://%s", options.URL)
			}
		}
	}

	options.Cookies, err = cmd.Flags().GetString("cookies")
	if err != nil {
		return options, fmt.Errorf("invalid value for cookies: %v", err)
	}

	options.Username, err = cmd.Flags().GetString("username")
	if err != nil {
		return options, fmt.Errorf("invalid value for username: %v", err)
	}

	options.Password, err = cmd.Flags().GetString("password")
	if err != nil {
		return options, fmt.Errorf("invalid value for password: %v", err)
	}

	options.UserAgent, err = cmd.Flags().GetString("useragent")
	if err != nil {
		return options, fmt.Errorf("invalid value for useragent: %v", err)
	}

	options.Proxy, err = cmd.Flags().GetString("proxy")
	if err != nil {
		return options, fmt.Errorf("invalid value for proxy: %v", err)
	}

	options.Timeout, err = cmd.Flags().GetDuration("timeout")
	if err != nil {
		return options, fmt.Errorf("invalid value for timeout: %v", err)
	}

	options.FollowRedirect, err = cmd.Flags().GetBool("followredirect")
	if err != nil {
		return options, fmt.Errorf("invalid value for followredirect: %v", err)
	}

	options.InsecureSSL, err = cmd.Flags().GetBool("insecuressl")
	if err != nil {
		return options, fmt.Errorf("invalid value for insecuressl: %v", err)
	}

	headers, err := cmd.Flags().GetStringArray("headers")
	if err != nil {
		return options, fmt.Errorf("invalid value for headers: %v", err)
	}

	for _, h := range headers {
		keyAndValue := strings.SplitN(h, ":", 2)
		if len(keyAndValue) != 2 {
			return options, fmt.Errorf("invalid header format for header %q", h)
		}
		key := strings.TrimSpace(keyAndValue[0])
		value := strings.TrimSpace(keyAndValue[1])
		if len(key) == 0 {
			return options, fmt.Errorf("invalid header format for header %q - name is empty", h)
		}
		header := libgobuster.HTTPHeader{Name: key, Value: value}
		options.Headers = append(options.Headers, header)
	}

	// Prompt for PW if not provided
	if options.Username != "" && options.Password == "" {
		fmt.Printf("[?] Auth Password: ")
		// please don't remove the int cast here as it is sadly needed on windows :/
		passBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		// print a newline to simulate the newline that was entered
		// this means that formatting/printing after doesn't look bad.
		fmt.Println("")
		if err != nil {
			return options, fmt.Errorf("username given but reading of password failed")
		}
		options.Password = string(passBytes)
	}
	// if it's still empty bail out
	if options.Username != "" && options.Password == "" {
		return options, fmt.Errorf("username was provided but password is missing")
	}

	return options, nil
}
