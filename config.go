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

// Contains config that are read in from the command
// line when the program is invoked.
type config struct {
	Client         *http.Client
	Cookies        string
	Expanded       bool
	Extensions     []string
	FollowRedirect bool
	IncludeLength  bool
	Mode           string
	NoStatus       bool
	Password       string
	Printer        func(cfg *config, br *busterResult)
	Processor      func(cfg *config, entity string, brc chan<- busterResult)
	ProxyUrl       *url.URL
	Quiet          bool
	Setup          func(cfg *config) bool
	ShowIPs        bool
	ShowCNAME      bool
	StatusCodes    statuscodes
	Threads        int
	Url            string
	UseSlash       bool
	UserAgent      string
	Username       string
	Verbose        bool
	Wordlist       string
	OutputFileName string
	OutputFile     *os.File
	IsWildcard     bool
	WildcardForced bool
	WildcardIps    ipwildcards
	SignalChan     chan os.Signal
	Terminate      bool
	StdIn          bool
	InsecureSSL    bool
}

// Parse all the command line options into a settings
// instance for future use.
func ParseCmdLine() *config {
	var extensions string
	var codes string
	var proxy string
	valid := true

	cfg := config{
		StatusCodes: statuscodes{sc: map[int]bool{}},
		WildcardIps: ipwildcards{ipw: map[string]bool{}},
		IsWildcard:  false,
		StdIn:       false,
	}

	// Set up the variables we're interested in parsing.
	flag.IntVar(&cfg.Threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&cfg.Mode, "m", "dir", "Directory/File mode (dir) or DNS mode (dns)")
	flag.StringVar(&cfg.Wordlist, "w", "", "Path to the wordlist")
	flag.StringVar(&codes, "s", "200,204,301,302,307", "Positive status codes (dir mode only)")
	flag.StringVar(&cfg.OutputFileName, "o", "", "Output file to write results to (defaults to stdout)")
	flag.StringVar(&cfg.Url, "u", "", "The target URL or Domain")
	flag.StringVar(&cfg.Cookies, "c", "", "Cookies to use for the requests (dir mode only)")
	flag.StringVar(&cfg.Username, "U", "", "Username for Basic Auth (dir mode only)")
	flag.StringVar(&cfg.Password, "P", "", "Password for Basic Auth (dir mode only)")
	flag.StringVar(&extensions, "x", "", "File extension(s) to search for (dir mode only)")
	flag.StringVar(&cfg.UserAgent, "a", "", "Set the User-Agent string (dir mode only)")
	flag.StringVar(&proxy, "p", "", "Proxy to use for requests [http(s)://host:port] (dir mode only)")
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose output (errors)")
	flag.BoolVar(&cfg.ShowIPs, "i", false, "Show IP addresses (dns mode only)")
	flag.BoolVar(&cfg.ShowCNAME, "cn", false, "Show CNAME records (dns mode only, cannot be used with '-i' option)")
	flag.BoolVar(&cfg.FollowRedirect, "r", false, "Follow redirects")
	flag.BoolVar(&cfg.Quiet, "q", false, "Don't print the banner and other noise")
	flag.BoolVar(&cfg.Expanded, "e", false, "Expanded mode, print full URLs")
	flag.BoolVar(&cfg.NoStatus, "n", false, "Don't print status codes")
	flag.BoolVar(&cfg.IncludeLength, "l", false, "Include the length of the body in the output (dir mode only)")
	flag.BoolVar(&cfg.UseSlash, "f", false, "Append a forward-slash to each directory request (dir mode only)")
	flag.BoolVar(&cfg.WildcardForced, "fw", false, "Force continued operation when wildcard found")
	flag.BoolVar(&cfg.InsecureSSL, "k", false, "Skip SSL certificate verification")

	flag.Parse()

	printBanner(&cfg)

	switch strings.ToLower(cfg.Mode) {
	case "dir":
		cfg.Printer = printDirResult
		cfg.Processor = processURL
		cfg.Setup = setupURL
	case "dns":
		cfg.Printer = printDnsResult
		cfg.Processor = processDNS
		cfg.Setup = setupDNS
	default:
		fmt.Println("[!] Mode (-m): Invalid value:", cfg.Mode)
		valid = false
	}

	if cfg.Threads < 0 {
		fmt.Println("[!] Threads (-t): Invalid value:", cfg.Threads)
		valid = false
	}

	stdin, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("[!] Unable to stat stdin, falling back to wordlist file.")
	} else if (stdin.Mode()&os.ModeCharDevice) == 0 && stdin.Size() > 0 {
		cfg.StdIn = true
	}

	if !cfg.StdIn {
		if cfg.Wordlist == "" {
			fmt.Println("[!] WordList (-w): Must be specified")
			valid = false
		} else if _, err := os.Stat(cfg.Wordlist); os.IsNotExist(err) {
			fmt.Println("[!] Wordlist (-w): File does not exist:", cfg.Wordlist)
			valid = false
		}
	} else if cfg.Wordlist != "" {
		fmt.Println("[!] Wordlist (-w) specified with pipe from stdin. Can't have both!")
		valid = false
	}

	if cfg.Url == "" {
		fmt.Println("[!] Url/Domain (-u): Must be specified")
		valid = false
	}

	if cfg.Mode == "dir" {
		if strings.HasSuffix(cfg.Url, "/") == false {
			cfg.Url = cfg.Url + "/"
		}

		if strings.HasPrefix(cfg.Url, "http") == false {
			// check to see if a port was specified
			re := regexp.MustCompile(`^[^/]+:(\d+)`)
			match := re.FindStringSubmatch(cfg.Url)

			if len(match) < 2 {
				// no port, default to http on 80
				cfg.Url = "http://" + cfg.Url
			} else {
				port, err := strconv.Atoi(match[1])
				if err != nil || (port != 80 && port != 443) {
					fmt.Println("[!] Url/Domain (-u): Scheme not specified.")
					valid = false
				} else if port == 80 {
					cfg.Url = "http://" + cfg.Url
				} else {
					cfg.Url = "https://" + cfg.Url
				}
			}
		}

		// extensions are comma separated
		if extensions != "" {
			cfg.Extensions = strings.Split(extensions, ",")
			for i := range cfg.Extensions {
				if cfg.Extensions[i][0] != '.' {
					cfg.Extensions[i] = "." + cfg.Extensions[i]
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
					cfg.StatusCodes.add(i)
				}
			}
		}

		// prompt for password if needed
		if valid && cfg.Username != "" && cfg.Password == "" {
			fmt.Printf("[?] Auth Password: ")
			passBytes, err := terminal.ReadPassword(int(syscall.Stdin))

			// print a newline to simulate the newline that was entered
			// this means that formatting/printing after doesn't look bad.
			fmt.Println("")

			if err == nil {
				cfg.Password = string(passBytes)
			} else {
				fmt.Println("[!] Auth username given but reading of password failed")
				valid = false
			}
		}

		if valid {
			var proxyUrlFunc func(*http.Request) (*url.URL, error)
			proxyUrlFunc = http.ProxyFromEnvironment

			if proxy != "" {
				proxyUrl, err := url.Parse(proxy)
				if err != nil {
					panic("[!] Proxy URL is invalid")
				}
				cfg.ProxyUrl = proxyUrl
				proxyUrlFunc = http.ProxyURL(cfg.ProxyUrl)
			}

			cfg.Client = &http.Client{
				Transport: &redirectHandler{
					Config: &cfg,
					Transport: &http.Transport{
						Proxy: proxyUrlFunc,
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: cfg.InsecureSSL,
						},
					},
				}}

			code, _ := get(&cfg, cfg.Url, "", cfg.Cookies)
			if code == nil {
				fmt.Println("[-] Unable to connect:", cfg.Url)
				valid = false
			}
		} else {
			printRuler(&cfg)
		}
	}

	if valid {
		return &cfg
	}

	return nil
}
