//----------------------------------------------------
// Gobuster -- by OJ Reeves
//
// A crap attempt at building something that resembles
// dirbuster or dirb using Go. The goal was to build
// a tool that would help learn Go and to actually do
// something useful. The idea of having this compile
// to native code is also appealing.
//
// Run: gobuster -h
//
// Please see THANKS file for contributors.
// Please see LICENSE file for license details.
//
//----------------------------------------------------

package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"

	"github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh/terminal"
)

// A single result which comes from an individual web
// request.
type Result struct {
	Entity string
	Status int
	Extra  string
	Size   *int64
}

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
	Printer        func(cfg *config, r *Result)
	Processor      func(cfg *config, entity string, resultChan chan<- Result)
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

// Small helper to combine URL with URI then make a
// request to the generated location.
func get(cfg *config, url, uri, cookie string) (*int, *int64) {
	return getResponse(cfg, url+uri, cookie)
}

// Make a request to the given URL.
func getResponse(cfg *config, fullUrl, cookie string) (*int, *int64) {
	req, err := http.NewRequest("GET", fullUrl, nil)

	if err != nil {
		return nil, nil
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	if cfg.UserAgent != "" {
		req.Header.Set("User-Agent", cfg.UserAgent)
	}

	if cfg.Username != "" {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}

	resp, err := cfg.Client.Do(req)

	if err != nil {
		if ue, ok := err.(*url.Error); ok {

			if strings.HasPrefix(ue.Err.Error(), "x509") {
				fmt.Println("[-] Invalid certificate")
			}

			if re, ok := ue.Err.(*redirectError); ok {
				return &re.StatusCode, nil
			}
		}
		return nil, nil
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("[!] problem closing the response body")
		}
	}()

	var length *int64 = nil

	if cfg.IncludeLength {
		length = new(int64)
		if resp.ContentLength <= 0 {
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				*length = int64(utf8.RuneCountInString(string(body)))
			}
		} else {
			*length = resp.ContentLength
		}
	}

	return &resp.StatusCode, length
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
		cfg.Processor = ProcessDirEntry
		cfg.Setup = SetupDir
	case "dns":
		cfg.Printer = printDnsResult
		cfg.Processor = ProcessDnsEntry
		cfg.Setup = SetupDns
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

// Process the busting of the website with the given
// set of settings from the command line.
func Process(cfg *config) {

	printConfig(cfg)

	if cfg.Setup(cfg) == false {
		printRuler(cfg)
		return
	}

	PrepareSignalHandler(cfg)

	// channels used for comms
	wordChan := make(chan string, cfg.Threads)
	resultChan := make(chan Result)

	// Use a wait group for waiting for all threads
	// to finish
	processorGroup := new(sync.WaitGroup)
	processorGroup.Add(cfg.Threads)
	printerGroup := new(sync.WaitGroup)
	printerGroup.Add(1)

	// Create goroutines for each of the number of threads
	// specified.
	for i := 0; i < cfg.Threads; i++ {
		go func() {
			for {
				word := <-wordChan

				// Did we reach the end? If so break.
				if word == "" {
					break
				}

				// Mode-specific processing
				cfg.Processor(cfg, word, resultChan)
			}

			// Indicate to the wait group that the thread
			// has finished.
			processorGroup.Done()
		}()
	}

	// Single goroutine which handles the results as they
	// appear from the worker threads.
	go func() {
		for r := range resultChan {
			cfg.Printer(cfg, &r)
		}
		printerGroup.Done()
	}()

	var scanner *bufio.Scanner

	if cfg.StdIn {
		// Read directly from stdin
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		// Pull content from the wordlist
		wordlist, err := os.Open(cfg.Wordlist)
		if err != nil {
			panic("Failed to open wordlist")
		}
		defer wordlist.Close()

		// Lazy reading of the wordlist line by line
		scanner = bufio.NewScanner(wordlist)
	}

	var outputFile *os.File
	if cfg.OutputFileName != "" {
		outputFile, err := os.Create(cfg.OutputFileName)
		if err != nil {
			fmt.Printf("[!] Unable to write to %s, falling back to stdout.\n", cfg.OutputFileName)
			cfg.OutputFileName = ""
			cfg.OutputFile = nil
		} else {
			cfg.OutputFile = outputFile
		}
	}

	for scanner.Scan() {
		if cfg.Terminate {
			break
		}
		word := strings.TrimSpace(scanner.Text())

		// Skip "comment" (starts with #), as well as empty lines
		if !strings.HasPrefix(word, "#") && len(word) > 0 {
			wordChan <- word
		}
	}

	close(wordChan)
	processorGroup.Wait()
	close(resultChan)
	printerGroup.Wait()
	if cfg.OutputFile != nil {
		outputFile.Close()
	}
	printRuler(cfg)
}

func SetupDns(cfg *config) bool {
	// Resolve a subdomain that probably shouldn't exist
	guid := uuid.NewV4()
	wildcardIps, err := net.LookupHost(fmt.Sprintf("%s.%s", guid, cfg.Url))
	if err == nil {
		cfg.IsWildcard = true
		cfg.WildcardIps.addRange(wildcardIps)
		fmt.Println("[-] Wildcard DNS found. IP address(es): ", cfg.WildcardIps.string())
		if !cfg.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard DNS, specify the '-fw' switch.")
		}
		return cfg.WildcardForced
	}

	if !cfg.Quiet {
		// Provide a warning if the base domain doesn't resolve (in case of typo)
		_, err = net.LookupHost(cfg.Url)
		if err != nil {
			// Not an error, just a warning. Eg. `yp.to` doesn't resolve, but `cr.py.to` does!
			fmt.Println("[-] Unable to validate base domain:", cfg.Url)
		}
	}

	return true
}

func SetupDir(cfg *config) bool {
	guid := uuid.NewV4()
	wildcardResp, _ := get(cfg, cfg.Url, fmt.Sprintf("%s", guid), cfg.Cookies)

	if cfg.StatusCodes.contains(*wildcardResp) {
		cfg.IsWildcard = true
		fmt.Println("[-] Wildcard response found:", fmt.Sprintf("%s%s", cfg.Url, guid), "=>", *wildcardResp)
		if !cfg.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard responses, specify the '-fw' switch.")
		}
		return cfg.WildcardForced
	}

	return true
}

func ProcessDnsEntry(cfg *config, word string, resultChan chan<- Result) {
	subdomain := word + "." + cfg.Url
	ips, err := net.LookupHost(subdomain)

	if err == nil {
		if !cfg.IsWildcard || !cfg.WildcardIps.containsAny(ips) {
			result := Result{
				Entity: subdomain,
			}
			if cfg.ShowIPs {
				result.Extra = strings.Join(ips, ", ")
			} else if cfg.ShowCNAME {
				cname, err := net.LookupCNAME(subdomain)
				if err == nil {
					result.Extra = cname
				}
			}
			resultChan <- result
		}
	} else if cfg.Verbose {
		result := Result{
			Entity: subdomain,
			Status: 404,
		}
		resultChan <- result
	}
}

func ProcessDirEntry(cfg *config, word string, resultChan chan<- Result) {
	suffix := ""
	if cfg.UseSlash {
		suffix = "/"
	}

	// Try the DIR first
	dirResp, dirSize := get(cfg, cfg.Url, word+suffix, cfg.Cookies)
	if dirResp != nil {
		resultChan <- Result{
			Entity: word + suffix,
			Status: *dirResp,
			Size:   dirSize,
		}
	}

	// Follow up with files using each ext.
	for ext := range cfg.Extensions {
		file := word + cfg.Extensions[ext]
		fileResp, fileSize := get(cfg, cfg.Url, file, cfg.Cookies)

		if fileResp != nil {
			resultChan <- Result{
				Entity: file,
				Status: *fileResp,
				Size:   fileSize,
			}
		}
	}
}

func WriteToFile(cfg *config, output string) {
	_, err := cfg.OutputFile.WriteString(output)
	if err != nil {
		log.Panicf("[!] Unable to write to file %v", cfg.OutputFileName)
	}
}

func PrepareSignalHandler(cfg *config) {
	cfg.SignalChan = make(chan os.Signal, 1)
	signal.Notify(cfg.SignalChan, os.Interrupt)
	go func() {
		for range cfg.SignalChan {
			// caught CTRL+C
			if !cfg.Quiet {
				fmt.Println("[!] Keyboard interrupt detected, terminating.")
				cfg.Terminate = true
			}
		}
	}()
}

func main() {
	state := ParseCmdLine()
	if state != nil {
		Process(state)
	}
}
