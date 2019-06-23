package libgobuster

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
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
	context          context.Context
	requestsExpected int
	requestsIssued   int
	mu               *sync.RWMutex
	plugin           GobusterPlugin
	resultChan       chan Result
	errorChan        chan error
	LogInfo          *log.Logger
	LogError         *log.Logger
}

// NewGobuster returns a new Gobuster object
func NewGobuster(c context.Context, opts *Options, plugin GobusterPlugin) (*Gobuster, error) {
	var g Gobuster
	g.Opts = opts
	g.plugin = plugin
	g.mu = new(sync.RWMutex)
	g.context = c
	g.resultChan = make(chan Result)
	g.errorChan = make(chan error)
	g.LogInfo = log.New(os.Stdout, "", log.LstdFlags)
	g.LogError = log.New(os.Stderr, "[ERROR] ", log.LstdFlags)

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

			wordCleaned := strings.TrimSpace(word)
			// Skip "comment" (starts with #), as well as empty lines
			if strings.HasPrefix(wordCleaned, "#") || len(wordCleaned) == 0 {
				break
			}

			// Mode-specific processing
			res, err := g.plugin.Run(wordCleaned)
			if err != nil {
				// do not exit and continue
				g.errorChan <- err
				continue
			} else {
				for _, r := range res {
					g.resultChan <- r
				}
			}

			select {
			case <-g.context.Done():
			case <-time.After(g.Opts.Delay):
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
	defer close(g.resultChan)
	defer close(g.errorChan)

	if err := g.plugin.PreRun(); err != nil {
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
		case wordChan <- scanner.Text():
		}
	}
	close(wordChan)
	workerGroup.Wait()
	return nil
}

// GetConfigString returns the current config as a printable string
func (g *Gobuster) GetConfigString() (string, error) {
	return g.plugin.GetConfigString()
}
