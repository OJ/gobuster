package libgobuster

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	// VERSION contains the current gobuster version
	VERSION = "1.4.1"
)

// SetupFunc is the "setup" function prototype for implementations
type SetupFunc func(*Gobuster) error

// ProcessFunc is the "process" function prototype for implementations
type ProcessFunc func(*Gobuster, string) ([]Result, error)

// ResultToStringFunc is the "to string" function prototype for implementations
type ResultToStringFunc func(*Gobuster, *Result) (*string, error)

// Gobuster is the main object when creating a new run
type Gobuster struct {
	Opts             *Options
	http             *httpClient
	WildcardIps      stringSet
	context          context.Context
	requestsExpected int
	requestsIssued   int
	mu               *sync.RWMutex
	funcResToString  ResultToStringFunc
	funcProcessor    ProcessFunc
	funcSetup        SetupFunc
	IsWildcard       bool
	resultChan       chan Result
}

// NewGobuster returns a new Gobuster object
func NewGobuster(c context.Context, opts *Options, setupFunc SetupFunc, processFunc ProcessFunc, resultFunc ResultToStringFunc) (*Gobuster, error) {
	// validate given options
	multiErr := opts.validate()
	if multiErr != nil {
		return nil, multiErr
	}

	var g Gobuster
	g.WildcardIps = stringSet{Set: map[string]bool{}}
	g.context = c
	g.Opts = opts
	h, err := newHTTPClient(c, opts)
	if err != nil {
		return nil, err
	}
	g.http = h

	g.funcSetup = setupFunc
	g.funcProcessor = processFunc
	g.funcResToString = resultFunc
	g.mu = new(sync.RWMutex)

	g.resultChan = make(chan Result)

	return &g, nil
}

// Results returns a channel of Results
func (g *Gobuster) Results() <-chan Result {
	return g.resultChan
}

func (g *Gobuster) incrementRequests() {
	g.mu.Lock()
	g.requestsIssued++
	g.mu.Unlock()
}

// PrintProgress outputs the current wordlist progress to stderr
func (g *Gobuster) PrintProgress() {
	g.mu.RLock()
	if g.Opts.Wordlist == "-" {
		fmt.Fprintf(os.Stderr, "\rProgress: %d", g.requestsIssued)
	} else {
		fmt.Fprintf(os.Stderr, "\rProgress: %d / %d", g.requestsIssued, g.requestsExpected)
	}
	g.mu.RUnlock()
}

// ClearProgress removes the last status line from stderr
func (g *Gobuster) ClearProgress() {
	fmt.Fprint(os.Stderr, "\r")
}

// GetRequest issues a GET request to the target and returns
// the status code, length and an error
func (g *Gobuster) GetRequest(url string) (*int, *int64, error) {
	return g.http.makeRequest(url, g.Opts.Cookies)
}

func (g *Gobuster) worker(wordChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-g.context.Done():
			return
		case word := <-wordChan:
			g.incrementRequests()
			// Mode-specific processing
			res, err := g.funcProcessor(g, word)
			if err != nil {
				// do not exit and continue
				log.Printf("ERROR on word %s: %v", word, err)
				continue
			} else {
				for _, r := range res {
					g.resultChan <- r
				}
			}
		}
	}
}

func (g *Gobuster) getWordlist() (*bufio.Scanner, error) {
	if g.Opts.Wordlist == "-" {
		// Read directly from stdin
		return bufio.NewScanner(os.Stdin), nil
	}
	// Pull content from the wordlist
	wordlist, err := os.Open(g.Opts.Wordlist)
	if err != nil {
		return nil, fmt.Errorf("failed to open wordlist: %v", err)
	}

	lines, err := lineCounter(wordlist)
	if err != nil {
		return nil, fmt.Errorf("failed to get number of lines: %v", err)
	}

	// mutiply by extensions to get the total number of requests
	if len(g.Opts.ExtensionsParsed) > 0 {
		lines = lines + (lines * len(g.Opts.ExtensionsParsed))
	}
	g.requestsExpected = lines

	// rewind wordlist
	_, err = wordlist.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to rewind wordlist: %v", err)
	}
	return bufio.NewScanner(wordlist), nil
}

// Start the busting of the website with the given
// set of settings from the command line.
func (g *Gobuster) Start() error {
	if err := g.funcSetup(g); err != nil {
		return err
	}

	var printerGroup sync.WaitGroup
	printerGroup.Add(1)
	var workerGroup sync.WaitGroup
	workerGroup.Add(g.Opts.Threads)

	wordChan := make(chan string, g.Opts.Threads)

	// Create goroutines for each of the number of threads
	// specified.
	for i := 0; i < g.Opts.Threads; i++ {
		go g.worker(wordChan, &workerGroup)
	}

	scanner, err := g.getWordlist()
	if err != nil {
		return err
	}

Scan:
	for scanner.Scan() {
		select {
		case <-g.context.Done():
			break Scan
		default:
			word := strings.TrimSpace(scanner.Text())
			// Skip "comment" (starts with #), as well as empty lines
			if !strings.HasPrefix(word, "#") && len(word) > 0 {
				wordChan <- word
			}
		}
	}
	close(wordChan)
	return nil
}

// ShowConfig prints the current config to the screen
func (g *Gobuster) ShowConfig() {
	o := g.Opts
	fmt.Printf("[+] Mode         : %s\n", o.Mode)
	fmt.Printf("[+] Url/Domain   : %s\n", o.URL)
	fmt.Printf("[+] Threads      : %d\n", o.Threads)

	wordlist := "stdin (pipe)"
	if o.Wordlist != "-" {
		wordlist = o.Wordlist
	}
	fmt.Printf("[+] Wordlist     : %s\n", wordlist)

	if o.Mode == ModeDir {
		fmt.Printf("[+] Status codes : %s\n", o.StatusCodesParsed.Stringify())

		if o.Proxy != "" {
			fmt.Printf("[+] Proxy        : %s\n", o.Proxy)
		}

		if o.Cookies != "" {
			fmt.Printf("[+] Cookies      : %s\n", o.Cookies)
		}

		if o.UserAgent != "" {
			fmt.Printf("[+] User Agent   : %s\n", o.UserAgent)
		}

		if o.IncludeLength {
			fmt.Printf("[+] Show length  : true\n")
		}

		if o.Username != "" {
			fmt.Printf("[+] Auth User    : %s\n", o.Username)
		}

		if len(o.Extensions) > 0 {
			fmt.Printf("[+] Extensions   : %s\n", strings.Join(o.ExtensionsParsed, ","))
		}

		if o.UseSlash {
			fmt.Printf("[+] Add Slash    : true\n")
		}

		if o.FollowRedirect {
			fmt.Printf("[+] Follow Redir : true\n")
		}

		if o.Expanded {
			fmt.Printf("[+] Expanded     : true\n")
		}

		if o.NoStatus {
			fmt.Printf("[+] No status    : true\n")
		}

		if o.Verbose {
			fmt.Printf("[+] Verbose      : true\n")
		}

		fmt.Printf("[+] Timeout      : %s\n", o.Timeout.String())
	}
}
