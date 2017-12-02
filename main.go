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

type ipWildcards struct {
	ipw map[string]bool
}

// Contains State that are read in from the command
// line when the program is invoked.
type State struct {
	Client         *http.Client
	Cookies        string
	Expanded       bool
	Extensions     []string
	FollowRedirect bool
	IncludeLength  bool
	Mode           string
	NoStatus       bool
	Password       string
	Printer        func(s *State, r *Result)
	Processor      func(s *State, entity string, resultChan chan<- Result)
	ProxyUrl       *url.URL
	Quiet          bool
	Setup          func(s *State) bool
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
	WildcardIps    ipWildcards
	SignalChan     chan os.Signal
	Terminate      bool
	StdIn          bool
	InsecureSSL    bool
}

type RedirectHandler struct {
	Transport http.RoundTripper
	State     *State
}

type RedirectError struct {
	StatusCode int
}

// Add an element to a set
func (set *ipWildcards) Add(s string) bool {
	_, found := set.ipw[s]
	set.ipw[s] = true
	return !found
}

// Add a list of elements to a set
func (set *ipWildcards) AddRange(ss []string) {
	for _, s := range ss {
		set.ipw[s] = true
	}
}

// Test if an element is in a set
func (set *ipWildcards) Contains(s string) bool {
	_, found := set.ipw[s]
	return found
}

// Check if any of the elements exist
func (set *ipWildcards) ContainsAny(ss []string) bool {
	for _, s := range ss {
		if set.ipw[s] {
			return true
		}
	}
	return false
}

// Stringify the set
func (set *ipWildcards) Stringify() string {
	values := []string{}
	for s := range set.ipw {
		values = append(values, s)
	}
	return strings.Join(values, ",")
}

// Make a request to the given URL.
func MakeRequest(s *State, fullUrl, cookie string) (*int, *int64) {
	req, err := http.NewRequest("GET", fullUrl, nil)

	if err != nil {
		return nil, nil
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	if s.UserAgent != "" {
		req.Header.Set("User-Agent", s.UserAgent)
	}

	if s.Username != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}

	resp, err := s.Client.Do(req)

	if err != nil {
		if ue, ok := err.(*url.Error); ok {

			if strings.HasPrefix(ue.Err.Error(), "x509") {
				fmt.Println("[-] Invalid certificate")
			}

			if re, ok := ue.Err.(*RedirectError); ok {
				return &re.StatusCode, nil
			}
		}
		return nil, nil
	}

	defer resp.Body.Close()

	var length *int64 = nil

	if s.IncludeLength {
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

// Small helper to combine URL with URI then make a
// request to the generated location.
func GoGet(s *State, url, uri, cookie string) (*int, *int64) {
	return MakeRequest(s, url+uri, cookie)
}

// Parse all the command line options into a settings
// instance for future use.
func ParseCmdLine() *State {
	var extensions string
	var codes string
	var proxy string
	valid := true

	s := State{
		StatusCodes: statuscodes{sc: map[int]bool{}},
		WildcardIps: ipWildcards{ipw: map[string]bool{}},
		IsWildcard:  false,
		StdIn:       false,
	}

	// Set up the variables we're interested in parsing.
	flag.IntVar(&s.Threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&s.Mode, "m", "dir", "Directory/File mode (dir) or DNS mode (dns)")
	flag.StringVar(&s.Wordlist, "w", "", "Path to the wordlist")
	flag.StringVar(&codes, "s", "200,204,301,302,307", "Positive status codes (dir mode only)")
	flag.StringVar(&s.OutputFileName, "o", "", "Output file to write results to (defaults to stdout)")
	flag.StringVar(&s.Url, "u", "", "The target URL or Domain")
	flag.StringVar(&s.Cookies, "c", "", "Cookies to use for the requests (dir mode only)")
	flag.StringVar(&s.Username, "U", "", "Username for Basic Auth (dir mode only)")
	flag.StringVar(&s.Password, "P", "", "Password for Basic Auth (dir mode only)")
	flag.StringVar(&extensions, "x", "", "File extension(s) to search for (dir mode only)")
	flag.StringVar(&s.UserAgent, "a", "", "Set the User-Agent string (dir mode only)")
	flag.StringVar(&proxy, "p", "", "Proxy to use for requests [http(s)://host:port] (dir mode only)")
	flag.BoolVar(&s.Verbose, "v", false, "Verbose output (errors)")
	flag.BoolVar(&s.ShowIPs, "i", false, "Show IP addresses (dns mode only)")
	flag.BoolVar(&s.ShowCNAME, "cn", false, "Show CNAME records (dns mode only, cannot be used with '-i' option)")
	flag.BoolVar(&s.FollowRedirect, "r", false, "Follow redirects")
	flag.BoolVar(&s.Quiet, "q", false, "Don't print the banner and other noise")
	flag.BoolVar(&s.Expanded, "e", false, "Expanded mode, print full URLs")
	flag.BoolVar(&s.NoStatus, "n", false, "Don't print status codes")
	flag.BoolVar(&s.IncludeLength, "l", false, "Include the length of the body in the output (dir mode only)")
	flag.BoolVar(&s.UseSlash, "f", false, "Append a forward-slash to each directory request (dir mode only)")
	flag.BoolVar(&s.WildcardForced, "fw", false, "Force continued operation when wildcard found")
	flag.BoolVar(&s.InsecureSSL, "k", false, "Skip SSL certificate verification")

	flag.Parse()

	Banner(&s)

	switch strings.ToLower(s.Mode) {
	case "dir":
		s.Printer = PrintDirResult
		s.Processor = ProcessDirEntry
		s.Setup = SetupDir
	case "dns":
		s.Printer = PrintDnsResult
		s.Processor = ProcessDnsEntry
		s.Setup = SetupDns
	default:
		fmt.Println("[!] Mode (-m): Invalid value:", s.Mode)
		valid = false
	}

	if s.Threads < 0 {
		fmt.Println("[!] Threads (-t): Invalid value:", s.Threads)
		valid = false
	}

	stdin, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("[!] Unable to stat stdin, falling back to wordlist file.")
	} else if (stdin.Mode()&os.ModeCharDevice) == 0 && stdin.Size() > 0 {
		s.StdIn = true
	}

	if !s.StdIn {
		if s.Wordlist == "" {
			fmt.Println("[!] WordList (-w): Must be specified")
			valid = false
		} else if _, err := os.Stat(s.Wordlist); os.IsNotExist(err) {
			fmt.Println("[!] Wordlist (-w): File does not exist:", s.Wordlist)
			valid = false
		}
	} else if s.Wordlist != "" {
		fmt.Println("[!] Wordlist (-w) specified with pipe from stdin. Can't have both!")
		valid = false
	}

	if s.Url == "" {
		fmt.Println("[!] Url/Domain (-u): Must be specified")
		valid = false
	}

	if s.Mode == "dir" {
		if strings.HasSuffix(s.Url, "/") == false {
			s.Url = s.Url + "/"
		}

		if strings.HasPrefix(s.Url, "http") == false {
			// check to see if a port was specified
			re := regexp.MustCompile(`^[^/]+:(\d+)`)
			match := re.FindStringSubmatch(s.Url)

			if len(match) < 2 {
				// no port, default to http on 80
				s.Url = "http://" + s.Url
			} else {
				port, err := strconv.Atoi(match[1])
				if err != nil || (port != 80 && port != 443) {
					fmt.Println("[!] Url/Domain (-u): Scheme not specified.")
					valid = false
				} else if port == 80 {
					s.Url = "http://" + s.Url
				} else {
					s.Url = "https://" + s.Url
				}
			}
		}

		// extensions are comma separated
		if extensions != "" {
			s.Extensions = strings.Split(extensions, ",")
			for i := range s.Extensions {
				if s.Extensions[i][0] != '.' {
					s.Extensions[i] = "." + s.Extensions[i]
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
					s.StatusCodes.add(i)
				}
			}
		}

		// prompt for password if needed
		if valid && s.Username != "" && s.Password == "" {
			fmt.Printf("[?] Auth Password: ")
			passBytes, err := terminal.ReadPassword(int(syscall.Stdin))

			// print a newline to simulate the newline that was entered
			// this means that formatting/printing after doesn't look bad.
			fmt.Println("")

			if err == nil {
				s.Password = string(passBytes)
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
				s.ProxyUrl = proxyUrl
				proxyUrlFunc = http.ProxyURL(s.ProxyUrl)
			}

			s.Client = &http.Client{
				Transport: &RedirectHandler{
					State: &s,
					Transport: &http.Transport{
						Proxy: proxyUrlFunc,
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: s.InsecureSSL,
						},
					},
				}}

			code, _ := GoGet(&s, s.Url, "", s.Cookies)
			if code == nil {
				fmt.Println("[-] Unable to connect:", s.Url)
				valid = false
			}
		} else {
			Ruler(&s)
		}
	}

	if valid {
		return &s
	}

	return nil
}

// Process the busting of the website with the given
// set of settings from the command line.
func Process(s *State) {

	ShowConfig(s)

	if s.Setup(s) == false {
		Ruler(s)
		return
	}

	PrepareSignalHandler(s)

	// channels used for comms
	wordChan := make(chan string, s.Threads)
	resultChan := make(chan Result)

	// Use a wait group for waiting for all threads
	// to finish
	processorGroup := new(sync.WaitGroup)
	processorGroup.Add(s.Threads)
	printerGroup := new(sync.WaitGroup)
	printerGroup.Add(1)

	// Create goroutines for each of the number of threads
	// specified.
	for i := 0; i < s.Threads; i++ {
		go func() {
			for {
				word := <-wordChan

				// Did we reach the end? If so break.
				if word == "" {
					break
				}

				// Mode-specific processing
				s.Processor(s, word, resultChan)
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
			s.Printer(s, &r)
		}
		printerGroup.Done()
	}()

	var scanner *bufio.Scanner

	if s.StdIn {
		// Read directly from stdin
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		// Pull content from the wordlist
		wordlist, err := os.Open(s.Wordlist)
		if err != nil {
			panic("Failed to open wordlist")
		}
		defer wordlist.Close()

		// Lazy reading of the wordlist line by line
		scanner = bufio.NewScanner(wordlist)
	}

	var outputFile *os.File
	if s.OutputFileName != "" {
		outputFile, err := os.Create(s.OutputFileName)
		if err != nil {
			fmt.Printf("[!] Unable to write to %s, falling back to stdout.\n", s.OutputFileName)
			s.OutputFileName = ""
			s.OutputFile = nil
		} else {
			s.OutputFile = outputFile
		}
	}

	for scanner.Scan() {
		if s.Terminate {
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
	if s.OutputFile != nil {
		outputFile.Close()
	}
	Ruler(s)
}

func SetupDns(s *State) bool {
	// Resolve a subdomain that probably shouldn't exist
	guid := uuid.NewV4()
	wildcardIps, err := net.LookupHost(fmt.Sprintf("%s.%s", guid, s.Url))
	if err == nil {
		s.IsWildcard = true
		s.WildcardIps.AddRange(wildcardIps)
		fmt.Println("[-] Wildcard DNS found. IP address(es): ", s.WildcardIps.Stringify())
		if !s.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard DNS, specify the '-fw' switch.")
		}
		return s.WildcardForced
	}

	if !s.Quiet {
		// Provide a warning if the base domain doesn't resolve (in case of typo)
		_, err = net.LookupHost(s.Url)
		if err != nil {
			// Not an error, just a warning. Eg. `yp.to` doesn't resolve, but `cr.py.to` does!
			fmt.Println("[-] Unable to validate base domain:", s.Url)
		}
	}

	return true
}

func SetupDir(s *State) bool {
	guid := uuid.NewV4()
	wildcardResp, _ := GoGet(s, s.Url, fmt.Sprintf("%s", guid), s.Cookies)

	if s.StatusCodes.contains(*wildcardResp) {
		s.IsWildcard = true
		fmt.Println("[-] Wildcard response found:", fmt.Sprintf("%s%s", s.Url, guid), "=>", *wildcardResp)
		if !s.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard responses, specify the '-fw' switch.")
		}
		return s.WildcardForced
	}

	return true
}

func ProcessDnsEntry(s *State, word string, resultChan chan<- Result) {
	subdomain := word + "." + s.Url
	ips, err := net.LookupHost(subdomain)

	if err == nil {
		if !s.IsWildcard || !s.WildcardIps.ContainsAny(ips) {
			result := Result{
				Entity: subdomain,
			}
			if s.ShowIPs {
				result.Extra = strings.Join(ips, ", ")
			} else if s.ShowCNAME {
				cname, err := net.LookupCNAME(subdomain)
				if err == nil {
					result.Extra = cname
				}
			}
			resultChan <- result
		}
	} else if s.Verbose {
		result := Result{
			Entity: subdomain,
			Status: 404,
		}
		resultChan <- result
	}
}

func ProcessDirEntry(s *State, word string, resultChan chan<- Result) {
	suffix := ""
	if s.UseSlash {
		suffix = "/"
	}

	// Try the DIR first
	dirResp, dirSize := GoGet(s, s.Url, word+suffix, s.Cookies)
	if dirResp != nil {
		resultChan <- Result{
			Entity: word + suffix,
			Status: *dirResp,
			Size:   dirSize,
		}
	}

	// Follow up with files using each ext.
	for ext := range s.Extensions {
		file := word + s.Extensions[ext]
		fileResp, fileSize := GoGet(s, s.Url, file, s.Cookies)

		if fileResp != nil {
			resultChan <- Result{
				Entity: file,
				Status: *fileResp,
				Size:   fileSize,
			}
		}
	}
}

func PrintDnsResult(s *State, r *Result) {
	output := ""
	if r.Status == 404 {
		output = fmt.Sprintf("Missing: %s\n", r.Entity)
	} else if s.ShowIPs {
		output = fmt.Sprintf("Found: %s [%s]\n", r.Entity, r.Extra)
	} else if s.ShowCNAME {
		output = fmt.Sprintf("Found: %s [%s]\n", r.Entity, r.Extra)
	} else {
		output = fmt.Sprintf("Found: %s\n", r.Entity)
	}
	fmt.Printf("%s", output)

	if s.OutputFile != nil {
		WriteToFile(output, s)
	}
}

func PrintDirResult(s *State, r *Result) {
	output := ""

	// Prefix if we're in verbose mode
	if s.Verbose {
		if s.StatusCodes.contains(r.Status) {
			output = "Found : "
		} else {
			output = "Missed: "
		}
	}

	if s.StatusCodes.contains(r.Status) || s.Verbose {
		if s.Expanded {
			output += s.Url
		} else {
			output += "/"
		}
		output += r.Entity

		if !s.NoStatus {
			output += fmt.Sprintf(" (Status: %d)", r.Status)
		}

		if r.Size != nil {
			output += fmt.Sprintf(" [Size: %d]", *r.Size)
		}
		output += "\n"

		fmt.Printf(output)

		if s.OutputFile != nil {
			WriteToFile(output, s)
		}
	}
}

func WriteToFile(output string, s *State) {
	_, err := s.OutputFile.WriteString(output)
	if err != nil {
		panic("[!] Unable to write to file " + s.OutputFileName)
	}
}

func PrepareSignalHandler(s *State) {
	s.SignalChan = make(chan os.Signal, 1)
	signal.Notify(s.SignalChan, os.Interrupt)
	go func() {
		for range s.SignalChan {
			// caught CTRL+C
			if !s.Quiet {
				fmt.Println("[!] Keyboard interrupt detected, terminating.")
				s.Terminate = true
			}
		}
	}()
}

func (e *RedirectError) Error() string {
	return fmt.Sprintf("Redirect code: %d", e.StatusCode)
}

func (rh *RedirectHandler) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if rh.State.FollowRedirect {
		return rh.Transport.RoundTrip(req)
	}

	resp, err = rh.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusNotModified, http.StatusUseProxy, http.StatusTemporaryRedirect:
		return nil, &RedirectError{StatusCode: resp.StatusCode}
	}

	return resp, err
}

func Ruler(s *State) {
	if !s.Quiet {
		fmt.Println("=====================================================")
	}
}

func Banner(state *State) {
	if state.Quiet {
		return
	}

	fmt.Println("")
	fmt.Println("Gobuster v1.3                OJ Reeves (@TheColonial)")
	Ruler(state)
}

func ShowConfig(state *State) {
	if state.Quiet {
		return
	}

	if state != nil {
		fmt.Printf("[+] Mode         : %s\n", state.Mode)
		fmt.Printf("[+] Url/Domain   : %s\n", state.Url)
		fmt.Printf("[+] Threads      : %d\n", state.Threads)

		wordlist := "stdin (pipe)"
		if !state.StdIn {
			wordlist = state.Wordlist
		}
		fmt.Printf("[+] Wordlist     : %s\n", wordlist)

		if state.OutputFileName != "" {
			fmt.Printf("[+] Output file  : %s\n", state.OutputFileName)
		}

		if state.Mode == "dir" {
			fmt.Printf("[+] Status codes : %s\n", state.StatusCodes.string())

			if state.ProxyUrl != nil {
				fmt.Printf("[+] Proxy        : %s\n", state.ProxyUrl)
			}

			if state.Cookies != "" {
				fmt.Printf("[+] Cookies      : %s\n", state.Cookies)
			}

			if state.UserAgent != "" {
				fmt.Printf("[+] User Agent   : %s\n", state.UserAgent)
			}

			if state.IncludeLength {
				fmt.Printf("[+] Show length  : true\n")
			}

			if state.Username != "" {
				fmt.Printf("[+] Auth User    : %s\n", state.Username)
			}

			if len(state.Extensions) > 0 {
				fmt.Printf("[+] Extensions   : %s\n", strings.Join(state.Extensions, ","))
			}

			if state.UseSlash {
				fmt.Printf("[+] Add Slash    : true\n")
			}

			if state.FollowRedirect {
				fmt.Printf("[+] Follow Redir : true\n")
			}

			if state.Expanded {
				fmt.Printf("[+] Expanded     : true\n")
			}

			if state.NoStatus {
				fmt.Printf("[+] No status    : true\n")
			}

			if state.Verbose {
				fmt.Printf("[+] Verbose      : true\n")
			}
		}

		Ruler(state)
	}
}

func main() {
	state := ParseCmdLine()
	if state != nil {
		Process(state)
	}
}
