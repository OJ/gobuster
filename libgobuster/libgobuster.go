package libgobuster

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
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

type Guess struct {
	word              string
	discoverOnSuccess bool
}

type Wordlist struct {
	scanner        *bufio.Scanner
	guessesPerLine int
	isStream       bool
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

func (g *Gobuster) worker(ctx context.Context, guessChan <-chan *Guess, successChan chan<- *Guess, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		// Prioritize stopping when the context is done
		select {
		case <-ctx.Done():
			return
		default:
		}
		select {
		case <-ctx.Done():
			return
		case guess := <-guessChan:

			// Mode-specific processing
			res, err := g.plugin.ProcessWord(ctx, guess.word, g.Progress)
			if err != nil {
				// do not exit and continue
				g.Progress.ErrorChan <- fmt.Errorf("error on word %s: %w", guess.word, err)
			}

			if res != nil {
				g.Progress.ResultChan <- res

				select {
				case <-ctx.Done():
					g.Progress.incrementRequests()
					return
				case successChan <- guess:
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

func feed(ctx context.Context, guessChan chan<- *Guess, words []string, discoverOnSuccess bool) {
	for _, w := range words {
		guess := &Guess{word: w, discoverOnSuccess: discoverOnSuccess}
		// Prioritize stopping when the context is done
		select {
		case <-ctx.Done():
			return
		default:
		}
		select {
		// need to check here too otherwise guessChan will block
		case <-ctx.Done():
			return
		case guessChan <- guess:
		}
	}
}

func (g *Gobuster) feeder(ctx context.Context, guessChan chan<- *Guess, words []string, discoverOnSuccess bool, wg *sync.WaitGroup) {
	defer wg.Done()

	feed(ctx, guessChan, words, discoverOnSuccess)
}

func (g *Gobuster) feedWordlist(ctx context.Context, guessChan chan<- *Guess, wordlist *Wordlist, wg *sync.WaitGroup) {
	defer wg.Done()

	for wordlist.scanner.Scan() {
		// Prioritize stopping when the context is done
		select {
		case <-ctx.Done():
			return
		default:
		}

		word := strings.TrimSpace(wordlist.scanner.Text())

		switch {
		case wordlist.isStream && len(word) != 0:
			// Increment to keep track of expected work
			g.Progress.IncrementTotalRequests(wordlist.guessesPerLine)
		case wordlist.isStream && len(word) == 0:
			// Skip empty lines without incrementing
			continue
		case len(word) == 0:
			// Skip empty lines removing expected work
			g.Progress.IncrementTotalRequests(-1 * wordlist.guessesPerLine)
			continue
		}

		if len(g.Opts.Patterns) > 0 {
			for _, w := range g.processPatterns(word) {
				guess := &Guess{word: w, discoverOnSuccess: true}
				select {
				case <-ctx.Done():
					return
				case guessChan <- guess:
				}

				feed(ctx, guessChan, g.plugin.AdditionalWords(w), true)
			}
		} else {
			guess := &Guess{word: word, discoverOnSuccess: true}

			select {
			case <-ctx.Done():
				return
			case guessChan <- guess:
			}

			feed(ctx, guessChan, g.plugin.AdditionalWords(word), true)
		}
	}
}

func (g *Gobuster) getWordlist(wordlist io.ReadSeeker) (*Wordlist, error) {
	// calculate expected requests
	var guessesPerLine int
	if len(g.Opts.Patterns) > 0 {
		nPats := len(g.Opts.Patterns)
		guessesPerLine = nPats + nPats*g.plugin.AdditionalWordsLen()
	} else {
		guessesPerLine = 1 + g.plugin.AdditionalWordsLen()
	}

	if g.Opts.Wordlist == "-" {
		// Read directly from stdin
		return &Wordlist{scanner: bufio.NewScanner(os.Stdin), guessesPerLine: guessesPerLine, isStream: true}, nil
	}

	lines, err := lineCounter(wordlist)
	if err != nil {
		return nil, fmt.Errorf("failed to get number of lines: %w", err)
	}

	if lines-g.Opts.WordlistOffset <= 0 {
		return nil, errors.New("offset is greater than the number of lines in the wordlist")
	}

	g.Progress.IncrementTotalRequests(lines * guessesPerLine)

	// add offset if needed (offset defaults to 0)
	g.Progress.incrementRequestsIssues(g.Opts.WordlistOffset * guessesPerLine)

	// rewind wordlist after lineCounter
	_, err = wordlist.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to rewind wordlist: %w", err)
	}

	wordlistScanner := bufio.NewScanner(wordlist)

	// skip lines
	for range g.Opts.WordlistOffset {
		if !wordlistScanner.Scan() {
			if err := wordlistScanner.Err(); err != nil {
				return nil, fmt.Errorf("failed to skip lines in wordlist: %w", err)
			}
			return nil, errors.New("failed to skip lines in wordlist")
		}
	}

	return &Wordlist{scanner: wordlistScanner, guessesPerLine: guessesPerLine, isStream: false}, nil
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

	guessChan := make(chan *Guess, g.Opts.Threads*3)
	successChan := make(chan *Guess)

	var f io.ReadSeekCloser
	if g.Opts.Wordlist != "-" { // stdin case is handled inside getWordlist
		var err error
		f, err = os.Open(g.Opts.Wordlist)
		if err != nil {
			return fmt.Errorf("failed to open wordlist: %w", err)
		}
		defer f.Close()
	}

	wordlist, err := g.getWordlist(f)
	if err != nil {
		return err
	}

	// Create goroutines for each of the number of threads
	// specified.
	for range g.Opts.Threads {
		go g.worker(workerCtx, guessChan, successChan, &workerGroup)
	}

	feederGroup.Add(1)
	go g.feedWordlist(feederCtx, guessChan, wordlist, &feederGroup)

ListenForMore:
	for {
		// Prioritize stopping when the context is done
		select {
		case <-ctx.Done():
			break ListenForMore
		default:
		}

		select {
		case <-ctx.Done():
			break ListenForMore
		case successGuess := <-successChan:
			// Add more guesses based on the results of previous attempts
			if successGuess.discoverOnSuccess {
				discoverWords := g.plugin.AdditionalSuccessWords(successGuess.word)
				if len(discoverWords) > 0 {
					g.Progress.IncrementTotalRequests(len(discoverWords))
					feederGroup.Add(1)
					go g.feeder(feederCtx, guessChan, discoverWords, false, &feederGroup)
				}

				patternDiscoverWords := g.processDiscoverPatterns(successGuess.word)
				if len(patternDiscoverWords) > 0 {
					g.Progress.IncrementTotalRequests(len(patternDiscoverWords))
					feederGroup.Add(1)
					go g.feeder(feederCtx, guessChan, patternDiscoverWords, false, &feederGroup)
				}
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

	if err := wordlist.scanner.Err(); err != nil {
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

	return g.applyPatterns(word, g.Opts.Patterns)
}

func (g *Gobuster) processDiscoverPatterns(word string) []string {
	if g.Opts.DiscoverPatternFile == "" {
		return nil
	}

	return g.applyPatterns(word, g.Opts.DiscoverPatterns)
}

func (g *Gobuster) applyPatterns(word string, patterns []string) []string {
	pat := make([]string, len(patterns))
	for i, x := range patterns {
		pat[i] = strings.ReplaceAll(x, PATTERN, word)
	}
	return pat
}
