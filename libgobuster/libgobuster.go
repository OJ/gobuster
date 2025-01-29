package libgobuster

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// PATTERN is the pattern for wordlist replacements in pattern file
const PATTERN = "{GOBUSTER}"

// SetupFunc is the "setup" function prototype for implementations
type SetupFunc func(*Gobuster) error

// ProcessFunc is the "process" function prototype for implementations
type ProcessFunc func(*Gobuster, string) ([]Result, error)

// ResultToStringFunc is the "to string" function prototype for implementations
type ResultToStringFunc func(*Gobuster, *Result) (*string, error)

// Gobuster is the main object when creating a new run
type Gobuster struct {
	Opts     *Options
	Logger   *Logger
	plugin   GobusterPlugin
	Progress *Progress
}

// NewGobuster returns a new Gobuster object
func NewGobuster(opts *Options, plugin GobusterPlugin, logger *Logger) (*Gobuster, error) {
	var g Gobuster
	g.Opts = opts
	g.plugin = plugin
	g.Logger = logger
	g.Progress = NewProgress()

	return &g, nil
}

func (g *Gobuster) worker(ctx context.Context, wordChan <-chan string, moreWordsChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case word := <-wordChan:

			wordCleaned := strings.TrimSpace(word)
			// Skip empty lines
			if len(wordCleaned) == 0 {
				g.Progress.incrementRequests()
				break
			}

			// Mode-specific processing
			res, err := g.plugin.ProcessWord(ctx, wordCleaned, g.Progress)
			if err != nil {
				// do not exit and continue
				g.Progress.ErrorChan <- fmt.Errorf("error on word %s: %w", wordCleaned, err)
			}

			if res != nil {
				g.Progress.ResultChan <- res
				select {
				case <-ctx.Done():
					return
				case moreWordsChan <- wordCleaned:
				}
			}

			g.Progress.incrementRequests()

			select {
			case <-ctx.Done():
			case <-time.After(g.Opts.Delay):
			}
		}
	}
}

func feed(ctx context.Context, wordChan chan<- string, words []string) {
	for _, w := range words {
		select {
		// need to check here too otherwise wordChan will block
		case <-ctx.Done():
			return
		case wordChan <- w:
		}
	}
}

func (g *Gobuster) feeder(ctx context.Context, wordChan chan<- string, words []string, wg *sync.WaitGroup) {
	defer wg.Done()

	feed(ctx, wordChan, words)
}

func (g *Gobuster) feedScanner(ctx context.Context, wordChan chan<- string, scanner *bufio.Scanner, wg *sync.WaitGroup) {
	defer wg.Done()

	for scanner.Scan() {
		word := scanner.Text()
		// add the original word
		select {
		case <-ctx.Done():
			return
		case wordChan <- word:
		}
		// now create perms
		for _, w := range g.processPatterns(word) {
			select {
			case <-ctx.Done():
				return
			case wordChan <- w:
			}

			feed(ctx, wordChan, g.plugin.AdditionalWords(w))
		}
		feed(ctx, wordChan, g.plugin.AdditionalWords(word))
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
		return nil, fmt.Errorf("failed to open wordlist: %w", err)
	}

	lines, err := lineCounter(wordlist)
	if err != nil {
		return nil, fmt.Errorf("failed to get number of lines: %w", err)
	}

	if lines-g.Opts.WordlistOffset <= 0 {
		return nil, fmt.Errorf("offset is greater than the number of lines in the wordlist")
	}

	// calcutate expected requests
	nPats := 1 + len(g.Opts.Patterns)
	requestsPerLine := nPats + nPats*g.plugin.AdditionalWordsLen()
	g.Progress.IncrementTotalRequests(lines * requestsPerLine)

	// add offset if needed (offset defaults to 0)
	g.Progress.incrementRequestsIssues(g.Opts.WordlistOffset * requestsPerLine)

	// rewind wordlist
	_, err = wordlist.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to rewind wordlist: %w", err)
	}

	wordlistScanner := bufio.NewScanner(wordlist)

	// skip lines
	for i := 0; i < g.Opts.WordlistOffset; i++ {
		if !wordlistScanner.Scan() {
			if err := wordlistScanner.Err(); err != nil {
				return nil, fmt.Errorf("failed to skip lines in wordlist: %w", err)
			}
			return nil, fmt.Errorf("failed to skip lines in wordlist")
		}
	}

	return wordlistScanner, nil
}

// Run the busting of the website with the given
// set of settings from the command line.
func (g *Gobuster) Run(ctx context.Context) error {
	defer close(g.Progress.ResultChan)
	defer close(g.Progress.ErrorChan)
	defer close(g.Progress.MessageChan)

	if err := g.plugin.PreRun(ctx, g.Progress); err != nil {
		return err
	}

	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	feederCtx, feederCancel := context.WithCancel(ctx)
	defer feederCancel()

	var workerGroup, feederGroup sync.WaitGroup
	workerGroup.Add(g.Opts.Threads)

	wordChan := make(chan string, g.Opts.Threads*3)
	moreWordsChan := make(chan string)

	scanner, err := g.getWordlist()
	if err != nil {
		return err
	}

	// Create goroutines for each of the number of threads
	// specified.
	for i := 0; i < g.Opts.Threads; i++ {
		go g.worker(workerCtx, wordChan, moreWordsChan, &workerGroup)
	}

	feederGroup.Add(1)
	go g.feedScanner(feederCtx, wordChan, scanner, &feederGroup)

ListenForMore:
	for {
		select {
		case <-ctx.Done():
			break ListenForMore
		case successWord := <-moreWordsChan:
			// Add more guesses based on the results of previous attempts
			// TODO: limit guess recursion depth somehow
			//  (eg index.html -> index.html~ -> index.html~~ -> index.html~~~ ...)
			// TODO: add the option for arbitrary patterns based on successful finds
			additionalWords := g.plugin.AdditionalSuccessWords(successWord)
			if len(additionalWords) > 0 {
				g.Progress.IncrementTotalRequests(len(additionalWords))
				feederGroup.Add(1)
				go g.feeder(feederCtx, wordChan, additionalWords, &feederGroup)
			}
		case <-time.After(200 * time.Millisecond):
			// With requests issued only after the results are synchronously
			// reported, this is well ordered without the timeout, however it would
			// exert a lot of lock pressure during the run to keep doing this in a
			// hot loop
			if g.Progress.RequestsExpected() == g.Progress.RequestsIssued() {
				// All the expected requests have completed, there is no pending or
				// in-progress work. If moreWordsChan was buffered we would need to
				// check it again here to ensure no pending work was added while we
				// acquired the locks
				break ListenForMore
			}
		}
	}

	feederCancel()
	workerCancel()
	feederGroup.Wait()
	workerGroup.Wait()

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// GetConfigString returns the current config as a printable string
func (g *Gobuster) GetConfigString() (string, error) {
	return g.plugin.GetConfigString()
}

func (g *Gobuster) processPatterns(word string) []string {
	if g.Opts.PatternFile == "" {
		return nil
	}

	//nolint:prealloc
	var pat []string
	for _, x := range g.Opts.Patterns {
		repl := strings.ReplaceAll(x, PATTERN, word)
		pat = append(pat, repl)
	}
	return pat
}
