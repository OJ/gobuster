package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/OJ/gobuster/v3/helper"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func addBasicHTTPOptions(cmd *cobra.Command) {
	cmd.Flags().StringP("useragent", "a", libgobuster.DefaultUserAgent(), "Set the User-Agent string")
	cmd.Flags().BoolP("random-agent", "", false, "Use a random User-Agent string")
	cmd.Flags().StringP("proxy", "", "", "Proxy to use for requests [http(s)://host:port]")
	cmd.Flags().DurationP("timeout", "", 10*time.Second, "HTTP Timeout")
	cmd.Flags().BoolP("no-tls-validation", "k", false, "Skip TLS certificate verification")
}

func addCommonHTTPOptions(cmd *cobra.Command) error {
	addBasicHTTPOptions(cmd)
	cmd.Flags().StringP("url", "u", "", "The target URL")
	cmd.Flags().StringP("cookies", "c", "", "Cookies to use for the requests")
	cmd.Flags().StringP("username", "U", "", "Username for Basic Auth")
	cmd.Flags().StringP("password", "P", "", "Password for Basic Auth")
	cmd.Flags().BoolP("follow-redirect", "r", false, "Follow redirects")
	cmd.Flags().StringArrayP("headers", "H", []string{""}, "Specify HTTP headers, -H 'Header1: val1' -H 'Header2: val2'")
	cmd.Flags().StringP("method", "m", "GET", "Use the following HTTP method")

	if err := cmd.MarkFlagRequired("url"); err != nil {
		return fmt.Errorf("error on marking flag as required: %w", err)
	}

	return nil
}

func parseBasicHTTPOptions(cmd *cobra.Command) (libgobuster.BasicHTTPOptions, error) {
	options := libgobuster.BasicHTTPOptions{}
	var err error

	options.UserAgent, err = cmd.Flags().GetString("useragent")
	if err != nil {
		return options, fmt.Errorf("invalid value for useragent: %w", err)
	}
	randomUA, err := cmd.Flags().GetBool("random-agent")
	if err != nil {
		return options, fmt.Errorf("invalid value for random-agent: %w", err)
	}
	if randomUA {
		ua, err := helper.GetRandomUserAgent()
		if err != nil {
			return options, err
		}
		options.UserAgent = ua
	}

	options.Proxy, err = cmd.Flags().GetString("proxy")
	if err != nil {
		return options, fmt.Errorf("invalid value for proxy: %w", err)
	}

	options.Timeout, err = cmd.Flags().GetDuration("timeout")
	if err != nil {
		return options, fmt.Errorf("invalid value for timeout: %w", err)
	}

	options.NoTLSValidation, err = cmd.Flags().GetBool("no-tls-validation")
	if err != nil {
		return options, fmt.Errorf("invalid value for no-tls-validation: %w", err)
	}
	return options, nil
}

func parseCommonHTTPOptions(cmd *cobra.Command) (libgobuster.HTTPOptions, error) {
	options := libgobuster.HTTPOptions{}
	var err error

	basic, err := parseBasicHTTPOptions(cmd)
	if err != nil {
		return options, err
	}
	options.Proxy = basic.Proxy
	options.Timeout = basic.Timeout
	options.UserAgent = basic.UserAgent
	options.NoTLSValidation = basic.NoTLSValidation

	options.URL, err = cmd.Flags().GetString("url")
	if err != nil {
		return options, fmt.Errorf("invalid value for url: %w", err)
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
		return options, fmt.Errorf("invalid value for cookies: %w", err)
	}

	options.Username, err = cmd.Flags().GetString("username")
	if err != nil {
		return options, fmt.Errorf("invalid value for username: %w", err)
	}

	options.Password, err = cmd.Flags().GetString("password")
	if err != nil {
		return options, fmt.Errorf("invalid value for password: %w", err)
	}

	options.FollowRedirect, err = cmd.Flags().GetBool("follow-redirect")
	if err != nil {
		return options, fmt.Errorf("invalid value for follow-redirect: %w", err)
	}

	options.Method, err = cmd.Flags().GetString("method")
	if err != nil {
		return options, fmt.Errorf("invalid value for method: %w", err)
	}

	headers, err := cmd.Flags().GetStringArray("headers")
	if err != nil {
		return options, fmt.Errorf("invalid value for headers: %w", err)
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
		passBytes, err := term.ReadPassword(int(syscall.Stdin)) //nolint:unconvert
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
