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

func (g *Gobuster) worker(ctx context.Context, wordChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case word, ok := <-wordChan:
			// worker finished
			if !ok {
				return
			}
			g.Progress.incrementRequests()

			wordCleaned := strings.TrimSpace(word)
			// Skip empty lines
			if len(wordCleaned) == 0 {
				break
			}

			// Mode-specific processing
			err := g.plugin.ProcessWord(ctx, wordCleaned, g.Progress)
			if err != nil {
				// do not exit and continue
				g.Progress.ErrorChan <- fmt.Errorf("error on word %s: %w", wordCleaned, err)
				continue
			}

			select {
			case <-ctx.Done():
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
	g.Progress.IncrementTotalRequests(lines)

	// add offset if needed (offset defaults to 0)
	g.Progress.incrementRequestsIssues(g.Opts.WordlistOffset)

	// call the function once with a dummy entry to receive the number
	// of custom words per wordlist word
	customWordsLen := len(g.plugin.AdditionalWords("dummy"))
	if customWordsLen > 0 {
		origExpected := g.Progress.RequestsExpected()
		inc := origExpected * customWordsLen
		g.Progress.IncrementTotalRequests(inc)
	}

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

	var workerGroup sync.WaitGroup
	workerGroup.Add(g.Opts.Threads)

	wordChan := make(chan string, g.Opts.Threads)

	// Create goroutines for each of the number of threads
	// specified.
	for i := 0; i < g.Opts.Threads; i++ {
		go g.worker(ctx, wordChan, &workerGroup)
	}

	scanner, err := g.getWordlist()
	if err != nil {
		return err
	}

Scan:
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			break Scan
		default:
			word := scanner.Text()
			perms := g.processPatterns(word)
			// add the original word
			wordChan <- word
			// now create perms
			for _, w := range perms {
				select {
				// need to check here too otherwise wordChan will block
				case <-ctx.Done():
					break Scan
				case wordChan <- w:
				}
			}

			for _, w := range g.plugin.AdditionalWords(word) {
				select {
				// need to check here too otherwise wordChan will block
				case <-ctx.Done():
					break Scan
				case wordChan <- w:
				}
			}
		}
	}
	close(wordChan)
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
