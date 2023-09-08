package cli

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"software.sslmate.com/src/go-pkcs12"
)

func BasicHTTPOptions() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "useragent", Aliases: []string{"a"}, Value: libgobuster.DefaultUserAgent(), Usage: "Set the User-Agent string"},
		&cli.BoolFlag{Name: "random-agent", Aliases: []string{"rua"}, Value: false, Usage: "Use a random User-Agent string"},
		&cli.StringFlag{Name: "proxy", Usage: "Proxy to use for requests [http(s)://host:port] or [socks5://host:port]"},
		&cli.DurationFlag{Name: "timeout", Aliases: []string{"to"}, Value: 10 * time.Second, Usage: "HTTP Timeout"},
		&cli.BoolFlag{Name: "no-tls-validation", Aliases: []string{"k"}, Value: false, Usage: "Skip TLS certificate verification"},
		&cli.BoolFlag{Name: "retry", Value: false, Usage: "Should retry on request timeout"},
		&cli.IntFlag{Name: "retry-attempts", Aliases: []string{"ra"}, Value: 3, Usage: "Times to retry on request timeout"},
		&cli.StringFlag{Name: "client-cert-pem", Aliases: []string{"ccp"}, Usage: "public key in PEM format for optional TLS client certificates]"},
		&cli.StringFlag{Name: "client-cert-pem-key", Aliases: []string{"ccpk"}, Usage: "private key in PEM format for optional TLS client certificates (this key needs to have no password)"},
		&cli.StringFlag{Name: "client-cert-p12", Aliases: []string{"ccp12"}, Usage: "a p12 file to use for options TLS client certificates"},
		&cli.StringFlag{Name: "client-cert-p12-password", Aliases: []string{"ccp12p"}, Usage: "the password to the p12 file"},
	}
}

func ParseBasicHTTPOptions(c *cli.Context) (libgobuster.BasicHTTPOptions, error) {
	var opts libgobuster.BasicHTTPOptions
	opts.UserAgent = c.String("useragent")
	randomUA := c.Bool("random-agent")
	if randomUA {
		ua, err := libgobuster.GetRandomUserAgent()
		if err != nil {
			return opts, err
		}
		opts.UserAgent = ua
	}
	opts.Proxy = c.String("proxy")
	opts.Timeout = c.Duration("timeout")
	opts.NoTLSValidation = c.Bool("no-tls-validation")
	opts.RetryOnTimeout = c.Bool("retry")
	opts.RetryAttempts = c.Int("retry-attempts")

	pemFile := c.String("client-cert-pem")
	pemKeyFile := c.String("client-cert-pem-key")
	p12File := c.String("client-cert-p12")
	p12Pass := c.String("client-cert-p12-password")

	if pemFile != "" && p12File != "" {
		return opts, fmt.Errorf("please supply either a pem or a p12, not both")
	}

	if pemFile != "" {
		cert, err := tls.LoadX509KeyPair(pemFile, pemKeyFile)
		if err != nil {
			return opts, fmt.Errorf("could not load supplied pem key: %w", err)
		}
		opts.TLSCertificate = &cert
	} else if p12File != "" {
		p12Content, err := os.ReadFile(p12File)
		if err != nil {
			return opts, fmt.Errorf("could not read p12 %s: %w", p12File, err)
		}
		privKey, pubKey, _, err := pkcs12.DecodeChain(p12Content, p12Pass)
		if err != nil {
			return opts, fmt.Errorf("could not load P12: %w", err)
		}
		opts.TLSCertificate = &tls.Certificate{
			Certificate: [][]byte{pubKey.Raw},
			PrivateKey:  privKey,
		}
	}

	return opts, nil
}

func CommonHTTPOptions() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, []cli.Flag{
		&cli.StringFlag{Name: "url", Aliases: []string{"u"}, Usage: "The target URL", Required: true},
		&cli.StringFlag{Name: "cookies", Aliases: []string{"c"}, Usage: "Cookies to use for the requests"},
		&cli.StringFlag{Name: "username", Aliases: []string{"U"}, Usage: "Username for Basic Auth"},
		&cli.StringFlag{Name: "password", Aliases: []string{"P"}, Usage: "Password for Basic Auth"},
		&cli.BoolFlag{Name: "follow-redirect", Aliases: []string{"r"}, Value: false, Usage: "Follow redirects"},
		&cli.StringSliceFlag{Name: "headers", Aliases: []string{"H"}, Usage: "Specify HTTP headers, -H 'Header1: val1' -H 'Header2: val2'"},
		&cli.BoolFlag{Name: "no-canonicalize-headers", Aliases: []string{"nch"}, Value: false, Usage: "Do not canonicalize HTTP header names. If set header names are sent as is"},
		&cli.StringFlag{Name: "method", Aliases: []string{"m"}, Value: "GET", Usage: "the password to the p12 file"},
	}...)
	flags = append(flags, BasicHTTPOptions()...)
	return flags
}

func ParseCommonHTTPOptions(c *cli.Context) (libgobuster.HTTPOptions, error) {
	var opts libgobuster.HTTPOptions
	basic, err := ParseBasicHTTPOptions(c)
	if err != nil {
		return opts, err
	}
	opts.BasicHTTPOptions = basic

	opts.URL = c.String("url")
	if !strings.HasPrefix(opts.URL, "http") {
		// check to see if a port was specified
		re := regexp.MustCompile(`^[^/]+:(\d+)`)
		match := re.FindStringSubmatch(opts.URL)

		if len(match) < 2 {
			// no port, default to http on 80
			opts.URL = fmt.Sprintf("http://%s", opts.URL)
		} else {
			port, err2 := strconv.Atoi(match[1])
			if err2 != nil || (port != 80 && port != 443) {
				return opts, fmt.Errorf("url scheme not specified")
			} else if port == 80 {
				opts.URL = fmt.Sprintf("http://%s", opts.URL)
			} else {
				opts.URL = fmt.Sprintf("https://%s", opts.URL)
			}
		}
	}

	opts.Cookies = c.String("cookies")
	opts.Username = c.String("username")
	opts.Password = c.String("password")

	// Prompt for PW if not provided
	if opts.Username != "" && opts.Password == "" {
		fmt.Printf("[?] Auth Password: ")
		// please don't remove the int cast here as it is sadly needed on windows :/
		passBytes, err := term.ReadPassword(int(syscall.Stdin)) //nolint:unconvert
		// print a newline to simulate the newline that was entered
		// this means that formatting/printing after doesn't look bad.
		fmt.Println("")
		if err != nil {
			return opts, fmt.Errorf("username given but reading of password failed")
		}
		opts.Password = string(passBytes)
	}
	// if it's still empty bail out
	if opts.Username != "" && opts.Password == "" {
		return opts, fmt.Errorf("username was provided but password is missing")
	}

	opts.FollowRedirect = c.Bool("follow-redirect")
	opts.NoCanonicalizeHeaders = c.Bool("no-canonicalize-headers")
	opts.Method = c.String("method")

	for _, h := range c.StringSlice("headers") {
		keyAndValue := strings.SplitN(h, ":", 2)
		if len(keyAndValue) != 2 {
			return opts, fmt.Errorf("invalid header format for header %q", h)
		}
		key := strings.TrimSpace(keyAndValue[0])
		value := strings.TrimSpace(keyAndValue[1])
		if len(key) == 0 {
			return opts, fmt.Errorf("invalid header format for header %q - name is empty", h)
		}
		header := libgobuster.HTTPHeader{Name: key, Value: value}
		opts.Headers = append(opts.Headers, header)
	}

	return opts, nil
}

func GlobalOptions() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "wordlist", Aliases: []string{"w"}, Usage: "Path to the wordlist. Set to - to use STDIN.", Required: true},
		&cli.DurationFlag{Name: "delay", Aliases: []string{"d"}, Usage: "Time each thread waits between requests (e.g. 1500ms)"},
		&cli.IntFlag{Name: "threads", Aliases: []string{"t"}, Value: 10, Usage: "Number of concurrent threads"},
		&cli.IntFlag{Name: "wordlist-offset", Aliases: []string{"wo"}, Value: 0, Usage: "Resume from a given position in the wordlist"},
		&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Output file to write results to (defaults to stdout)"},
		&cli.BoolFlag{Name: "quiet", Aliases: []string{"q"}, Value: false, Usage: "Don't print the banner and other noise"},
		&cli.BoolFlag{Name: "no-progress", Aliases: []string{"np"}, Value: false, Usage: "Don't display progress"},
		&cli.BoolFlag{Name: "no-error", Aliases: []string{"ne"}, Value: false, Usage: "Don't display errors"},
		&cli.StringFlag{Name: "pattern", Aliases: []string{"p"}, Usage: "File containing replacement patterns"},
		&cli.BoolFlag{Name: "no-color", Aliases: []string{"nc"}, Value: false, Usage: "Disable color output"},
		&cli.BoolFlag{Name: "debug", Value: false, Usage: "enable debug output"},
	}
}

func ParseGlobalOptions(c *cli.Context) (libgobuster.Options, error) {
	var opts libgobuster.Options

	opts.Wordlist = c.String("wordlist")
	if opts.Wordlist == "-" {
		// STDIN
	} else if _, err := os.Stat(opts.Wordlist); os.IsNotExist(err) {
		return opts, fmt.Errorf("wordlist file %q does not exist: %w", opts.Wordlist, err)
	}

	opts.Delay = c.Duration("delay")
	opts.Threads = c.Int("threads")
	opts.WordlistOffset = c.Int("wordlist-offset")
	if opts.Wordlist == "-" && opts.WordlistOffset > 0 {
		return opts, fmt.Errorf("wordlist-offset is not supported when reading from STDIN")
	} else if opts.WordlistOffset < 0 {
		return opts, fmt.Errorf("wordlist-offset must be bigger or equal to 0")
	}

	opts.OutputFilename = c.String("output")
	opts.Quiet = c.Bool("quiet")
	opts.NoProgress = c.Bool("no-progress")
	opts.NoError = c.Bool("no-error")
	opts.PatternFile = c.String("pattern")
	if opts.PatternFile != "" {
		if _, err := os.Stat(opts.PatternFile); os.IsNotExist(err) {
			return opts, fmt.Errorf("pattern file %q does not exist: %w", opts.PatternFile, err)
		}
		patternFile, err := os.Open(opts.PatternFile)
		if err != nil {
			return opts, fmt.Errorf("could not open pattern file %q: %w", opts.PatternFile, err)
		}
		defer patternFile.Close()

		scanner := bufio.NewScanner(patternFile)
		for scanner.Scan() {
			opts.Patterns = append(opts.Patterns, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return opts, fmt.Errorf("could not read pattern file %q: %w", opts.PatternFile, err)
		}
	}

	if c.Bool("no-color") {
		color.NoColor = true
	}

	opts.Debug = c.Bool("debug")
	return opts, nil
}
