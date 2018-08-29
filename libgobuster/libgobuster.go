package libgobuster

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

const (
	// VERSION contains the current gobuster version
	VERSION = "2.0.0"
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
	plugin           GobusterPlugin
	IsWildcard       bool
	resultChan       chan Result
	errorChan        chan error
}

// GobusterPlugin is an interface which plugins must implement
type GobusterPlugin interface {
	Setup(*Gobuster) error
	Process(*Gobuster, string) ([]Result, error)
	ResultToString(*Gobuster, *Result) (*string, error)
}

// NewGobuster returns a new Gobuster object
func NewGobuster(c context.Context, opts *Options, plugin GobusterPlugin) (*Gobuster, error) {
	// validate given options
	multiErr := opts.validate()
	if multiErr != nil {
		return nil, multiErr
	}

	var g Gobuster
	g.WildcardIps = newStringSet()
	g.context = c
	g.Opts = opts
	h, err := newHTTPClient(c, opts)
	if err != nil {
		return nil, err
	}
	g.http = h

	g.plugin = plugin
	g.mu = new(sync.RWMutex)

	g.resultChan = make(chan Result)
	g.errorChan = make(chan error)

	return &g, nil
}

// Results returns a channel of Results
func (g *Gobuster) Results() <-chan Result {
	return g.resultChan
}

// Errors returns a channel of errors
func (g *Gobuster) Errors() <-chan error {
	return g.errorChan
}

func (g *Gobuster) incrementRequests() {
	g.mu.Lock()
	g.requestsIssued++
	g.mu.Unlock()
}

// PrintProgress outputs the current wordlist progress to stderr
func (g *Gobuster) PrintProgress() {
	if !g.Opts.Quiet && !g.Opts.NoProgress {
		g.mu.RLock()
		if g.Opts.Wordlist == "-" {
			fmt.Fprintf(os.Stderr, "\rProgress: %d", g.requestsIssued)
			// only print status if we already read in the wordlist
		} else if g.requestsExpected > 0 {
			fmt.Fprintf(os.Stderr, "\rProgress: %d / %d (%3.2f%%)", g.requestsIssued, g.requestsExpected, float32(g.requestsIssued)*100.0/float32(g.requestsExpected))
		}
		g.mu.RUnlock()
	}
}

// ClearProgress removes the last status line from stderr
func (g *Gobuster) ClearProgress() {
	fmt.Fprint(os.Stderr, resetTerminal())
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
		case word, ok := <-wordChan:
			// worker finished
			if !ok {
				return
			}
			g.incrementRequests()
			// Mode-specific processing
			res, err := g.plugin.Process(g, word)
			if err != nil {
				// do not exit and continue
				g.errorChan <- err
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
	if len(g.Opts.ExtensionsParsed.Set) > 0 {
		lines = lines + (lines * len(g.Opts.ExtensionsParsed.Set))
	}
	g.requestsExpected = lines
	g.requestsIssued = 0

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
	if err := g.plugin.Setup(g); err != nil {
		return err
	}

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
	workerGroup.Wait()
	close(g.resultChan)
	close(g.errorChan)
	return nil
}

// GetConfigString returns the current config as a printable string
func (g *Gobuster) GetConfigString() (string, error) {
	buf := &bytes.Buffer{}
	o := g.Opts
	if _, err := fmt.Fprintf(buf, "[+] Mode         : %s\n", o.Mode); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(buf, "[+] Url/Domain   : %s\n", o.URL); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(buf, "[+] Threads      : %d\n", o.Threads); err != nil {
		return "", err
	}

	wordlist := "stdin (pipe)"
	if o.Wordlist != "-" {
		wordlist = o.Wordlist
	}
	if _, err := fmt.Fprintf(buf, "[+] Wordlist     : %s\n", wordlist); err != nil {
		return "", err
	}

	if o.Mode == ModeDir {
		if _, err := fmt.Fprintf(buf, "[+] Status codes : %s\n", o.StatusCodesParsed.Stringify()); err != nil {
			return "", err
		}

		if o.Proxy != "" {
			if _, err := fmt.Fprintf(buf, "[+] Proxy        : %s\n", o.Proxy); err != nil {
				return "", err
			}
		}

		if o.Cookies != "" {
			if _, err := fmt.Fprintf(buf, "[+] Cookies      : %s\n", o.Cookies); err != nil {
				return "", err
			}
		}

		if o.UserAgent != "" {
			if _, err := fmt.Fprintf(buf, "[+] User Agent   : %s\n", o.UserAgent); err != nil {
				return "", err
			}
		}

		if o.IncludeLength {
			if _, err := fmt.Fprintf(buf, "[+] Show length  : true\n"); err != nil {
				return "", err
			}
		}

		if o.Username != "" {
			if _, err := fmt.Fprintf(buf, "[+] Auth User    : %s\n", o.Username); err != nil {
				return "", err
			}
		}

		if len(o.Extensions) > 0 {
			if _, err := fmt.Fprintf(buf, "[+] Extensions   : %s\n", o.ExtensionsParsed.Stringify()); err != nil {
				return "", err
			}
		}

		if o.UseSlash {
			if _, err := fmt.Fprintf(buf, "[+] Add Slash    : true\n"); err != nil {
				return "", err
			}
		}

		if o.FollowRedirect {
			if _, err := fmt.Fprintf(buf, "[+] Follow Redir : true\n"); err != nil {
				return "", err
			}
		}

		if o.Expanded {
			if _, err := fmt.Fprintf(buf, "[+] Expanded     : true\n"); err != nil {
				return "", err
			}
		}

		if o.NoStatus {
			if _, err := fmt.Fprintf(buf, "[+] No status    : true\n"); err != nil {
				return "", err
			}
		}

		if o.Verbose {
			if _, err := fmt.Fprintf(buf, "[+] Verbose      : true\n"); err != nil {
				return "", err
			}
		}

		if _, err := fmt.Fprintf(buf, "[+] Timeout      : %s\n", o.Timeout.String()); err != nil {
			return "", err
		}
	}

	return strings.TrimSpace(buf.String()), nil
}
