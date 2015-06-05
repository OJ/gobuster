package main

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
//----------------------------------------------------

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

// A single result which comes from an individual web
// request.
type Result struct {
	Entity string
	Status int
	Extra  string
}

type PrintResultFunc func(s *State, r *Result)
type ProcessorFunc func(s *State, entity string, resultChan chan<- Result)

// Shim type for "set"
type IntSet struct {
	set map[int]bool
}

// Contains State that are read in from the command
// line when the program is invoked.
type State struct {
	Threads        int
	Wordlist       string
	Url            string
	Cookies        string
	Extensions     []string
	StatusCodes    IntSet
	Verbose        bool
	UseSlash       bool
	FollowRedirect bool
	Quiet          bool
	Mode           string
	Printer        PrintResultFunc
	Processor      ProcessorFunc
	Client         *http.Client
}

type RedirectHandler struct {
	Transport http.RoundTripper
	State     *State
}

type RedirectError struct {
	StatusCode int
}

// Add an element to a set
func (set *IntSet) Add(i int) bool {
	_, found := set.set[i]
	set.set[i] = true
	return !found
}

// Test if an element is in a set
func (set *IntSet) Contains(i int) bool {
	_, found := set.set[i]
	return found
}

// Stringify the set
func (set *IntSet) Stringify() string {
	values := []string{}
	for s, _ := range set.set {
		values = append(values, strconv.Itoa(s))
	}
	return strings.Join(values, ",")
}

// Make a request to the given URL.
func MakeRequest(client *http.Client, fullUrl, cookie string) *int {
	req, err := http.NewRequest("GET", fullUrl, nil)

	if err != nil {
		return nil
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := client.Do(req)

	if err != nil {
		if ue, ok := err.(*url.Error); ok {
			if re, ok := ue.Err.(*RedirectError); ok {
				return &re.StatusCode
			}
		}
		return nil
	}

	defer resp.Body.Close()

	return &resp.StatusCode
}

// Small helper to combine URL with URI then make a
// request to the generated location.
func GoGet(client *http.Client, url, uri, cookie string) *int {
	return MakeRequest(client, url+uri, cookie)
}

// Parse all the command line options into a settings
// instance for future use.
func ParseCmdLine() *State {
	var extensions string
	var codes string
	valid := true

	s := State{StatusCodes: IntSet{set: map[int]bool{}}}

	// Set up the variables we're interested in parsing.
	flag.IntVar(&s.Threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&s.Mode, "m", "dir", "Directory/File mode (dir) or DNS mode (dns)")
	flag.StringVar(&s.Wordlist, "w", "", "Path to the wordlist")
	flag.StringVar(&codes, "s", "200,204,301,302,307", "Positive status codes (dir mode only)")
	flag.StringVar(&s.Url, "u", "", "The target URL or Domain")
	flag.StringVar(&s.Cookies, "c", "", "Cookies to use for the requests (dir mode only)")
	flag.StringVar(&extensions, "x", "", "File extension(s) to search for (dir mode only)")
	flag.BoolVar(&s.Verbose, "v", false, "Verbose output (errors and IP addresses")
	flag.BoolVar(&s.FollowRedirect, "r", false, "Follow redirects")
	flag.BoolVar(&s.Quiet, "q", false, "Only print found items.")
	flag.BoolVar(&s.UseSlash, "f", false, "Append a forward-slash to each directory request (dir mode only)")

	flag.Parse()

	switch strings.ToLower(s.Mode) {
	case "dir":
		s.Printer = PrintDirResult
		s.Processor = ProcessDirEntry
	case "dns":
		s.Printer = PrintDnsResult
		s.Processor = ProcessDnsEntry
	default:
		fmt.Println("Mode (-m): Invalid value:", s.Mode)
		valid = false
	}

	if s.Threads < 0 {
		fmt.Println("Threads (-t): Invalid value:", s.Threads)
		valid = false
	}

	if s.Wordlist == "" {
		fmt.Println("WordList (-w): Must be specified")
		valid = false
	} else if _, err := os.Stat(s.Wordlist); os.IsNotExist(err) {
		fmt.Println("Wordlist (-w): File does not exist:", s.Wordlist)
		valid = false
	}

	if s.Url == "" {
		fmt.Println("Url/Domain (-u): Must be specified")
		valid = false
	}

	if s.Mode == "dir" {
		if strings.HasSuffix(s.Url, "/") == false {
			s.Url = s.Url + "/"
		}

		// extensions are comma seaprated
		if extensions != "" {
			s.Extensions = strings.Split(extensions, ",")
		}

		// status codes are comma seaprated
		if codes != "" {
			for _, c := range strings.Split(codes, ",") {
				i, err := strconv.Atoi(c)
				if err != nil {
					panic("Invalid status code given")
				}
				s.StatusCodes.Add(i)
			}
		}

		if valid {
			s.Client = &http.Client{
				Transport: &RedirectHandler{
					State: &s,
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				}}

			if GoGet(s.Client, s.Url, "", s.Cookies) == nil {
				fmt.Println("[-] Unable to connect:", s.Url)
				valid = false
			}
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
	wordlist, err := os.Open(s.Wordlist)
	if err != nil {
		panic("Failed to open wordlist")
	}

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

	defer wordlist.Close()

	// Lazy reading of the wordlist line by line
	scanner := bufio.NewScanner(wordlist)
	for scanner.Scan() {
		word := scanner.Text()

		// Skip "comment" lines
		if strings.HasPrefix(word, "#") == false {
			wordChan <- word
		}
	}

	close(wordChan)
	processorGroup.Wait()
	close(resultChan)
	printerGroup.Wait()
}

func ProcessDnsEntry(s *State, word string, resultChan chan<- Result) {
	subdomain := word + "." + s.Url
	ips, err := net.LookupHost(subdomain)

	if err == nil {
		result := Result{
			Entity: subdomain,
		}
		if s.Verbose {
			result.Extra = strings.Join(ips, ", ")
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
	dirResp := GoGet(s.Client, s.Url, word+suffix, s.Cookies)
	if dirResp != nil {
		resultChan <- Result{
			Entity: word + suffix,
			Status: *dirResp,
		}
	}

	// Follow up with files using each ext.
	for ext := range s.Extensions {
		file := word + s.Extensions[ext]
		fileResp := GoGet(s.Client, s.Url, file, s.Cookies)
		if fileResp != nil {
			resultChan <- Result{
				Entity: file,
				Status: *fileResp,
			}
		}
	}
}

func PrintDnsResult(s *State, r *Result) {
	if s.Verbose {
		fmt.Printf("Found: %s [%s]\n", r.Entity, r.Extra)
	} else {
		fmt.Printf("Found: %s\n", r.Entity)
	}
}

func PrintDirResult(s *State, r *Result) {
	if s.StatusCodes.Contains(r.Status) {
		// Only print results out if we find something
		// meaningful.
		if s.Quiet {
			fmt.Printf("%s%s\n", s.Url, r.Entity)
		} else {
			fmt.Printf("Found: /%s (%d)\n", r.Entity, r.Status)
		}
	} else if s.Verbose {
		// Print out other results if the user wants to
		// see them.
		fmt.Printf("Result: /%s (%d)\n", r.Entity, r.Status)
	}
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

func Banner(state *State) {
	fmt.Println("\n=====================================================")
	fmt.Println("Gobuster v0.7 (DIR support by OJ Reeves @TheColonial)")
	fmt.Println("              (DNS support by Peleus     @0x42424242)")
	fmt.Println("=====================================================")

	if state != nil {
		fmt.Printf("[+] Mode         : %s\n", state.Mode)
		fmt.Printf("[+] Url/Domain   : %s\n", state.Url)
		fmt.Printf("[+] Threads      : %d\n", state.Threads)
		fmt.Printf("[+] Wordlist     : %s\n", state.Wordlist)

		if state.Mode == "dir" {
			fmt.Printf("[+] Status codes : %s\n", state.StatusCodes.Stringify())

			if state.Cookies != "" {
				fmt.Printf("[+] Cookies      : %s\n", state.Cookies)
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

			if state.Verbose {
				fmt.Printf("[+] Verbose      : true\n")
			}
		}
		fmt.Println("=====================================================")
	}
}

func main() {
	state := ParseCmdLine()
	if state.Quiet {
		Process(state)
	} else {
		Banner(state)
		Process(state)
		fmt.Println("=====================================================")
	}

}
