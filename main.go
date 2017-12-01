package main

//----------------------------------------------------
// Gobuster -- by OJ Reeves
//
// A crap attempt at building something that resembles dirbuster or dirb using
// Go. The goal was to build a tool that would help learn Go and to actually
// do something useful. The idea of having this compile to native code is also
// appealing.
//
// Run: gobuster -h
//
// Please see THANKS file for contributors.
// Please see LICENSE file for license details.
//
//----------------------------------------------------

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

const (
	dirMode = "dir"
	dnsMode = "dns"
)

// result is the response of an individual web request
type result struct {
	Entity string
	Status int
	Extra  string
	Size   *int64
}

type printResultFunc func(s *state, r *result)
type processorFunc func(s *state, entity string, resultChan chan<- result)
type setupFunc func(s *state) bool

// intSet is a shim type for "set" containing ints
type intSet struct {
	set map[int]bool
}

// stringSet is a shim type for "set" containing strings
type stringSet struct {
	set map[string]bool
}

// state contains data read in from the command line when the program is invoked
type state struct {
	Extensions     []string
	URL            string
	Cookies        string
	UserAgent      string
	Username       string
	Wordlist       string
	Mode           string
	OutputFileName string
	Password       string
	Printer        printResultFunc
	Processor      processorFunc
	ProxyURL       *url.URL
	OutputFile     *os.File
	Setup          setupFunc
	WildcardIps    stringSet
	SignalChan     chan os.Signal
	StatusCodes    intSet
	Threads        int
	Client         *http.Client
	UseSlash       bool
	ShowCNAME      bool
	ShowIPs        bool
	Verbose        bool
	Quiet          bool
	NoStatus       bool
	IncludeLength  bool
	IsWildcard     bool
	WildcardForced bool
	FollowRedirect bool
	Expanded       bool
	Terminate      bool
	StdIn          bool
	InsecureSSL    bool
}

type redirectHandler struct {
	Transport http.RoundTripper
	State     *state
}

type redirectError struct {
	StatusCode int
}

// addRange adds a list of elements to a set
func (set *stringSet) addRange(ss []string) {
	for _, s := range ss {
		set.set[s] = true
	}
}

// containsAny checks if any of the elements exist
func (set *stringSet) containsAny(ss []string) bool {
	for _, s := range ss {
		if set.set[s] {
			return true
		}
	}
	return false
}

// stringify returns the set as a string
func (set *stringSet) stringify() string {
	values := []string{}
	for s := range set.set {
		values = append(values, s)
	}
	return strings.Join(values, ",")
}

// add adds an element to a set
func (set *intSet) add(i int) bool {
	_, found := set.set[i]
	set.set[i] = true
	return !found
}

// contains tests if an element is in a set
func (set *intSet) contains(i int) bool {
	_, found := set.set[i]
	return found
}

// stringify returns the set as a string
func (set *intSet) stringify() string {
	values := []string{}
	for s := range set.set {
		values = append(values, strconv.Itoa(s))
	}
	return strings.Join(values, ",")
}

// makeRequest makes a request to the given URL
func makeRequest(s *state, fullURL, cookie string) (*int, *int64) {
	req, err := http.NewRequest("GET", fullURL, nil)

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

			if re, ok := ue.Err.(*redirectError); ok {
				return &re.StatusCode, nil
			}
		}
		return nil, nil
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	var length *int64

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

// goGet is a small helper to combine URL with URI then make a request to the generated location
func goGet(s *state, url, uri, cookie string) (*int, *int64) {
	return makeRequest(s, url+uri, cookie)
}

// parseCmdLine all the command line options into a settings instance for future use.
func parseCmdLine() *state {
	var extensions string
	var codes string
	var proxy string
	valid := true

	s := state{
		StatusCodes: intSet{set: map[int]bool{}},
		WildcardIps: stringSet{set: map[string]bool{}},
		IsWildcard:  false,
		StdIn:       false,
	}

	// Set up the variables we're interested in parsing.
	flag.IntVar(&s.Threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&s.Mode, "m", dirMode, "Directory/File mode (dirMode) or DNS mode (dnsMode)")
	flag.StringVar(&s.Wordlist, "w", "", "Path to the wordlist")
	flag.StringVar(&codes, "s", "200,204,301,302,307", "Positive status codes (dirMode only)")
	flag.StringVar(&s.OutputFileName, "o", "", "Output file to write results to (defaults to stdout)")
	flag.StringVar(&s.URL, "u", "", "The target URL or Domain")
	flag.StringVar(&s.Cookies, "c", "", "Cookies to use for the requests (dirMode only)")
	flag.StringVar(&s.Username, "U", "", "Username for Basic Auth (dirMode only)")
	flag.StringVar(&s.Password, "P", "", "Password for Basic Auth (dirMode only)")
	flag.StringVar(&extensions, "x", "", "File extension(s) to search for (dirMode only)")
	flag.StringVar(&s.UserAgent, "a", "", "Set the User-Agent string (dirMode only)")
	flag.StringVar(&proxy, "p", "", "Proxy to use for requests [http(s)://host:port] (dirMode only)")
	flag.BoolVar(&s.Verbose, "v", false, "Verbose output (errors)")
	flag.BoolVar(&s.ShowIPs, "i", false, "Show IP addresses (dnsMode only)")
	flag.BoolVar(&s.ShowCNAME, "cn", false, "Show CNAME records (dnsMode only, cannot be used with '-i' option)")
	flag.BoolVar(&s.FollowRedirect, "r", false, "Follow redirects")
	flag.BoolVar(&s.Quiet, "q", false, "Don't print the banner and other noise")
	flag.BoolVar(&s.Expanded, "e", false, "Expanded mode, print full URLs")
	flag.BoolVar(&s.NoStatus, "n", false, "Don't print status codes")
	flag.BoolVar(&s.IncludeLength, "l", false, "Include the length of the body in the output (dirMode only)")
	flag.BoolVar(&s.UseSlash, "f", false, "Append a forward-slash to each directory request (dirMode only)")
	flag.BoolVar(&s.WildcardForced, "fw", false, "Force continued operation when wildcard found")
	flag.BoolVar(&s.InsecureSSL, "k", false, "Skip SSL certificate verification")

	flag.Parse()

	banner(&s)

	switch strings.ToLower(s.Mode) {
	case dirMode:
		s.Printer = printDirResult
		s.Processor = processDirEntry
		s.Setup = setupDir
	case dnsMode:
		s.Printer = printDNSResult
		s.Processor = processDNSEntry
		s.Setup = setupDNS
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

	if s.URL == "" {
		fmt.Println("[!] URL/Domain (-u): Must be specified")
		valid = false
	}

	if s.Mode == dirMode {
		if !strings.HasSuffix(s.URL, "/") {
			s.URL = s.URL + "/"
		}

		if !strings.HasPrefix(s.URL, "http") {
			// check to see if a port was specified
			re := regexp.MustCompile(`^[^/]+:(\d+)`)
			match := re.FindStringSubmatch(s.URL)

			if len(match) < 2 {
				// no port, default to http on 80
				s.URL = "http://" + s.URL
			} else {
				port, err := strconv.Atoi(match[1])
				if err != nil || (port != 80 && port != 443) {
					fmt.Println("[!] URL/Domain (-u): Scheme not specified.")
					valid = false
				} else if port == 80 {
					s.URL = "http://" + s.URL
				} else {
					s.URL = "https://" + s.URL
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
			passBytes, err := terminal.ReadPassword(syscall.Stdin)

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
			proxyURL := http.ProxyFromEnvironment

			if proxy != "" {
				p, err := url.Parse(proxy)
				if err != nil {
					panic("[!] Proxy URL is invalid")
				}
				s.ProxyURL = p
				proxyURL = http.ProxyURL(s.ProxyURL)
			}

			s.Client = &http.Client{
				Transport: &redirectHandler{
					State: &s,
					Transport: &http.Transport{
						Proxy: proxyURL,
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: s.InsecureSSL, // nolint: gas
						},
					},
				}}

			code, _ := goGet(&s, s.URL, "", s.Cookies)
			if code == nil {
				fmt.Println("[-] Unable to connect:", s.URL)
				valid = false
			}
		} else {
			ruler(&s)
		}
	}

	if valid {
		return &s
	}

	return nil
}

// process the busting of the website with the given set of settings from the command line.
func process(s *state) {

	showConfig(s)

	if !s.Setup(s) {
		ruler(s)
		return
	}

	prepareSignalHandler(s)

	// channels used for comms
	wordChan := make(chan string, s.Threads)
	resultChan := make(chan result)

	// Use a wait group for waiting for all threads to finish
	processorGroup := new(sync.WaitGroup)
	processorGroup.Add(s.Threads)
	printerGroup := new(sync.WaitGroup)
	printerGroup.Add(1)

	// Create goroutines for each of the number of threads specified.
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

			// Indicate to the wait group that the thread has finished.
			processorGroup.Done()
		}()
	}

	// Single goroutine which handles the results as they appear from the worker threads.
	go func() {
		for r := range resultChan {
			s.Printer(s, &r)
		}
		printerGroup.Done()
	}()

	var scanner *bufio.Scanner

	if s.StdIn {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		wordlist, err := os.Open(s.Wordlist)
		if err != nil {
			panic("Failed to open wordlist")
		}
		defer func() {
			if err := wordlist.Close(); err != nil {
				log.Print(err)
			}
		}()

		scanner = bufio.NewScanner(wordlist)
	}

	var outputFile *os.File
	var err error
	if s.OutputFileName != "" {
		outputFile, err = os.Create(s.OutputFileName)
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
		if err := outputFile.Close(); err != nil {
			log.Print(err)
		}
	}
	ruler(s)
}

func setupDNS(s *state) bool {
	// Resolve a subdomain that probably shouldn't exist
	guid := uuid.NewV4()
	wildcardIps, err := net.LookupHost(fmt.Sprintf("%s.%s", guid, s.URL))
	if err == nil {
		s.IsWildcard = true
		s.WildcardIps.addRange(wildcardIps)
		fmt.Println("[-] Wildcard DNS found. IP address(es): ", s.WildcardIps.stringify())
		if !s.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard DNS, specify the '-fw' switch.")
		}
		return s.WildcardForced
	}

	if !s.Quiet {
		// Provide a warning if the base domain doesn't resolve (in case of typo)
		_, err = net.LookupHost(s.URL)
		if err != nil {
			// Not an error, just a warning. Eg. `yp.to` doesn't resolve, but `cr.py.to` does!
			fmt.Println("[-] Unable to validate base domain:", s.URL)
		}
	}

	return true
}

func setupDir(s *state) bool {
	guid := uuid.NewV4()
	wildcardResp, _ := goGet(s, s.URL, guid.String(), s.Cookies)

	if s.StatusCodes.contains(*wildcardResp) {
		s.IsWildcard = true
		fmt.Println("[-] Wildcard response found:", fmt.Sprintf("%s%s", s.URL, guid), "=>", *wildcardResp)
		if !s.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard responses, specify the '-fw' switch.")
		}
		return s.WildcardForced
	}

	return true
}

func processDNSEntry(s *state, word string, resultChan chan<- result) {
	subdomain := word + "." + s.URL
	ips, err := net.LookupHost(subdomain)

	if err == nil {
		if !s.IsWildcard || !s.WildcardIps.containsAny(ips) {
			r := result{
				Entity: subdomain,
			}
			if s.ShowIPs {
				r.Extra = strings.Join(ips, ", ")
			} else if s.ShowCNAME {
				cname, err := net.LookupCNAME(subdomain)
				if err == nil {
					r.Extra = cname
				}
			}
			resultChan <- r
		}
	} else if s.Verbose {
		r := result{
			Entity: subdomain,
			Status: 404,
		}
		resultChan <- r
	}
}

func processDirEntry(s *state, word string, resultChan chan<- result) {
	suffix := ""
	if s.UseSlash {
		suffix = "/"
	}

	// Try the DIR first
	dirResp, dirSize := goGet(s, s.URL, word+suffix, s.Cookies)
	if dirResp != nil {
		resultChan <- result{
			Entity: word + suffix,
			Status: *dirResp,
			Size:   dirSize,
		}
	}

	// Follow up with files using each ext.
	for ext := range s.Extensions {
		file := word + s.Extensions[ext]
		fileResp, fileSize := goGet(s, s.URL, file, s.Cookies)

		if fileResp != nil {
			resultChan <- result{
				Entity: file,
				Status: *fileResp,
				Size:   fileSize,
			}
		}
	}
}

func printDNSResult(s *state, r *result) {
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
		writeToFile(output, s)
	}
}

func printDirResult(s *state, r *result) {
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
			output += s.URL
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

		fmt.Print(output)

		if s.OutputFile != nil {
			writeToFile(output, s)
		}
	}
}

func writeToFile(output string, s *state) {
	_, err := s.OutputFile.WriteString(output)
	if err != nil {
		panic("[!] Unable to write to file " + s.OutputFileName)
	}
}

func prepareSignalHandler(s *state) {
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

func (e *redirectError) Error() string {
	return fmt.Sprintf("Redirect code: %d", e.StatusCode)
}

func (rh *redirectHandler) RoundTrip(req *http.Request) (resp *http.Response, err error) {
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
		return nil, &redirectError{StatusCode: resp.StatusCode}
	}

	return resp, err
}

func ruler(s *state) {
	if !s.Quiet {
		fmt.Println("=====================================================")
	}
}

func banner(state *state) {
	if state.Quiet {
		return
	}

	fmt.Println("")
	fmt.Println("Gobuster v1.3                OJ Reeves (@TheColonial)")
	ruler(state)
}

func showConfig(state *state) {
	if state.Quiet {
		return
	}

	if state != nil {
		fmt.Printf("[+] Mode         : %s\n", state.Mode)
		fmt.Printf("[+] URL/Domain   : %s\n", state.URL)
		fmt.Printf("[+] Threads      : %d\n", state.Threads)

		wordlist := "stdin (pipe)"
		if !state.StdIn {
			wordlist = state.Wordlist
		}
		fmt.Printf("[+] Wordlist     : %s\n", wordlist)

		if state.OutputFileName != "" {
			fmt.Printf("[+] Output file  : %s\n", state.OutputFileName)
		}

		if state.Mode == dirMode {
			fmt.Printf("[+] Status codes : %s\n", state.StatusCodes.stringify())

			if state.ProxyURL != nil {
				fmt.Printf("[+] Proxy        : %s\n", state.ProxyURL)
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
				fmt.Printf("[+] add Slash    : true\n")
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

		ruler(state)
	}
}

func main() {
	state := parseCmdLine()
	if state != nil {
		process(state)
	}
}
