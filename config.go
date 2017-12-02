// Config

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	modeURL = "dir"
	modeDNS = "dns"
)

// Contains config that are read in from the command
// line when the program is invoked.
type config struct { //nolint: maligned
	client         *http.Client
	cookies        string
	expanded       bool
	extensions     []string
	followRedirect bool
	includeLength  bool
	mode           string
	noStatus       bool
	password       string
	printer        func(cfg *config, br *busterResult)
	processor      func(cfg *config, entity string, brc chan<- busterResult)
	proxyURL       *url.URL
	quiet          bool
	setup          func(cfg *config) bool
	showIPs        bool
	showCNAME      bool
	statusCodes    statuscodes
	threads        int
	url            string
	useSlash       bool
	userAgent      string
	username       string
	verbose        bool
	wordlist       string
	outputFileName string
	outputFile     *os.File
	isWildcard     bool
	wildcardForced bool
	wildcardIps    ipwildcards
	signalChan     chan os.Signal
	terminate      bool
	stdIn          bool
	insecureSSL    bool
}

// Parse all the command line options into a settings
// instance for future use.
func parseCmdLine() *config { //nolint: gocyclo
	var extensions string
	var codes string
	var proxy string
	valid := true

	cfg := config{
		statusCodes: statuscodes{sc: map[int]bool{}},
		wildcardIps: ipwildcards{ipw: map[string]bool{}},
		isWildcard:  false,
		stdIn:       false,
	}

	// Set up the variables we're interested in parsing.
	flag.IntVar(&cfg.threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&cfg.mode, "m", "dir", "Directory/File mode (dir) or DNS mode (dns)")
	flag.StringVar(&cfg.wordlist, "w", "", "Path to the wordlist")
	flag.StringVar(&codes, "s", "200,204,301,302,307", "Positive status codes (dir mode only)")
	flag.StringVar(&cfg.outputFileName, "o", "", "Output file to write results to (defaults to stdout)")
	flag.StringVar(&cfg.url, "u", "", "The target URL or Domain")
	flag.StringVar(&cfg.cookies, "c", "", "Cookies to use for the requests (dir mode only)")
	flag.StringVar(&cfg.username, "U", "", "Username for Basic Auth (dir mode only)")
	flag.StringVar(&cfg.password, "P", "", "Password for Basic Auth (dir mode only)")
	flag.StringVar(&extensions, "x", "", "File extension(s) to search for (dir mode only)")
	flag.StringVar(&cfg.userAgent, "a", "", "Set the User-Agent string (dir mode only)")
	flag.StringVar(&proxy, "p", "", "Proxy to use for requests [http(s)://host:port] (dir mode only)")
	flag.BoolVar(&cfg.verbose, "v", false, "Verbose output (errors)")
	flag.BoolVar(&cfg.showIPs, "i", false, "Show IP addresses (dns mode only)")
	flag.BoolVar(&cfg.showCNAME, "cn", false, "Show CNAME records (dns mode only, cannot be used with '-i' option)")
	flag.BoolVar(&cfg.followRedirect, "r", false, "Follow redirects")
	flag.BoolVar(&cfg.quiet, "q", false, "Don't print the banner and other noise")
	flag.BoolVar(&cfg.expanded, "e", false, "Expanded mode, print full URLs")
	flag.BoolVar(&cfg.noStatus, "n", false, "Don't print status codes")
	flag.BoolVar(&cfg.includeLength, "l", false, "Include the length of the body in the output (dir mode only)")
	flag.BoolVar(&cfg.useSlash, "f", false, "Append a forward-slash to each directory request (dir mode only)")
	flag.BoolVar(&cfg.wildcardForced, "fw", false, "Force continued operation when wildcard found")
	flag.BoolVar(&cfg.insecureSSL, "k", false, "Skip SSL certificate verification")

	flag.Parse()

	printBanner(&cfg)

	switch strings.ToLower(cfg.mode) {
	case modeDNS:
		cfg.printer = printDNSResult
		cfg.processor = processDNS
		cfg.setup = setupDNS
	case modeURL:
		cfg.printer = printDirResult
		cfg.processor = processURL
		cfg.setup = setupURL
		if !strings.HasSuffix(cfg.url, "/") {
			cfg.url = cfg.url + "/"
		}

		if !strings.HasPrefix(cfg.url, "http") {
			// check to see if a port was specified
			re := regexp.MustCompile(`^[^/]+:(\d+)`)
			match := re.FindStringSubmatch(cfg.url)

			if len(match) < 2 {
				// no port, default to http on 80
				cfg.url = "http://" + cfg.url
			} else {
				port, err := strconv.Atoi(match[1])
				if err != nil || (port != 80 && port != 443) {
					fmt.Println("[!] Url/Domain (-u): Scheme not specified.")
					valid = false
				} else if port == 80 {
					cfg.url = "http://" + cfg.url
				} else {
					cfg.url = "https://" + cfg.url
				}
			}
		}

		// extensions are comma separated
		if extensions != "" {
			cfg.extensions = strings.Split(extensions, ",")
			for i := range cfg.extensions {
				if cfg.extensions[i][0] != '.' {
					cfg.extensions[i] = "." + cfg.extensions[i]
				}
			}
		}

		// status codes are comma separated
		if codes != "" {
			for _, c := range strings.Split(codes, ",") {
				i, err := strconv.Atoi(c)
				if err != nil {
					fmt.Println("[!] Invalid status code given: ", c)
					valid = false
				} else {
					cfg.statusCodes.add(i)
				}
			}
		}

		// prompt for password if needed
		if valid && cfg.username != "" && cfg.password == "" {
			fmt.Printf("[?] Auth Password: ")
			passBytes, err := terminal.ReadPassword(syscall.Stdin)

			// print a newline to simulate the newline that was entered
			// this means that formatting/printing after doesn't look bad.
			fmt.Println("")

			if err == nil {
				cfg.password = string(passBytes)
			} else {
				fmt.Println("[!] Auth username given but reading of password failed")
				valid = false
			}
		}

		if valid {
			proxyURLFunc := http.ProxyFromEnvironment

			if proxy != "" {
				proxyURL, err := url.Parse(proxy)
				if err != nil {
					panic("[!] Proxy URL is invalid")
				}
				cfg.proxyURL = proxyURL
				proxyURLFunc = http.ProxyURL(cfg.proxyURL)
			}

			cfg.client = &http.Client{
				Transport: &redirectHandler{
					Config: &cfg,
					Transport: &http.Transport{
						Proxy: proxyURLFunc,
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: cfg.insecureSSL, //nolint: gas
						},
					},
				}}

			code, _ := get(&cfg, cfg.url, "", cfg.cookies)
			if code == nil {
				fmt.Println("[-] Unable to connect:", cfg.url)
				valid = false
			}
		} else {
			printRuler(&cfg)
		}
	default:
		fmt.Println("[!] Mode (-m): Invalid value:", cfg.mode)
		valid = false
	}

	if cfg.threads < 0 {
		fmt.Println("[!] Threads (-t): Invalid value:", cfg.threads)
		valid = false
	}

	stdin, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("[!] Unable to stat stdin, falling back to wordlist file.")
	} else if (stdin.Mode()&os.ModeCharDevice) == 0 && stdin.Size() > 0 {
		cfg.stdIn = true
	}

	if !cfg.stdIn {
		if cfg.wordlist == "" {
			fmt.Println("[!] WordList (-w): Must be specified")
			valid = false
		} else if _, err := os.Stat(cfg.wordlist); os.IsNotExist(err) {
			fmt.Println("[!] Wordlist (-w): File does not exist:", cfg.wordlist)
			valid = false
		}
	} else if cfg.wordlist != "" {
		fmt.Println("[!] Wordlist (-w) specified with pipe from stdin. Can't have both!")
		valid = false
	}

	if cfg.url == "" {
		fmt.Println("[!] Url/Domain (-u): Must be specified")
		valid = false
	}

	if valid {
		return &cfg
	}

	return nil
}
