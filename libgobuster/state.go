package libgobuster

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// A single result which comes from an individual web
// request.
type Result struct {
	Entity string
	Status int
	Extra  string
	Size   *int64
}

type PrintResultFunc func(s *State, r *Result)
type ProcessorFunc func(s *State, entity string, resultChan chan<- Result)
type SetupFunc func(s *State) bool

// Shim type for "set" containing ints
type IntSet struct {
	Set map[int]bool
}

// Shim type for "set" containing strings
type StringSet struct {
	Set map[string]bool
}

// Contains State that are read in from the command
// line when the program is invoked.
type State struct {
	Client           *http.Client
	Cookies          string
	Expanded         bool
	Extensions       []string
	FollowRedirect   bool
	IncludeLength    bool
	Mode             string
	NoStatus         bool
	Password         string
	Printer          PrintResultFunc
	Processor        ProcessorFunc
	ProxyURL         *url.URL
	Quiet            bool
	Setup            SetupFunc
	ShowIPs          bool
	ShowCNAME        bool
	StatusCodes      IntSet
	Threads          int
	URL              string
	UseSlash         bool
	UserAgent        string
	Username         string
	Verbose          bool
	Wordlist         string
	OutputFileName   string
	OutputFile       *os.File
	IsWildcard       bool
	WildcardForced   bool
	WildcardIps      StringSet
	Context          context.Context
	StdIn            bool
	InsecureSSL      bool
	WordlistSize     int
	WordlistPosition int
	Mu               sync.RWMutex
	Timeout          time.Duration
}

// Process the busting of the website with the given
// set of settings from the command line.
func Process(s *State) {

	ShowConfig(s)

	if s.Setup(s) == false {
		Ruler(s)
		return
	}

	// channels used for comms
	wordChan := make(chan string, s.Threads)
	resultChan := make(chan Result)
	quitChan := make(chan int)

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
			// Indicate to the wait group that the thread
			// has finished.
			defer processorGroup.Done()
			for {
				select {
				case word := <-wordChan:
					// Did we reach the end? If so return.
					if word == "" {
						return
					}
					// Mode-specific processing
					s.Processor(s, word, resultChan)
				case <-s.Context.Done():
					return
				}
			}
		}()
	}

	// Single goroutine which handles the results as they
	// appear from the worker threads.
	go func() {
		tick := time.Tick(1 * time.Second)
		defer printerGroup.Done()
		for {
			select {
			case <-s.Context.Done():
				return
			case <-quitChan:
				// remove last status output
				fmt.Printf("\r")
				return
			case r := <-resultChan:
				s.Printer(s, &r)
			case <-tick:
				s.printStatus()
			}
		}
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

		lines, err := lineCounter(wordlist)
		if err != nil {
			panic("Failed to get number of lines")
		}
		s.WordlistSize = lines

		// rewind wordlist
		_, err = wordlist.Seek(0, 0)
		if err != nil {
			panic("Failed to rewind wordlist")
		}

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
		select {
		case <-s.Context.Done():
			break
		default:
			word := strings.TrimSpace(scanner.Text())
			// Skip "comment" (starts with #), as well as empty lines
			if !strings.HasPrefix(word, "#") && len(word) > 0 {
				wordChan <- word
			}
		}
	}
	close(wordChan)

	processorGroup.Wait()
	close(resultChan)
	close(quitChan)
	printerGroup.Wait()
	if s.OutputFile != nil {
		outputFile.Close()
	}
	Ruler(s)
}

func (s *State) printStatus() {
	s.Mu.RLock()
	fmt.Printf("\rStatus: %d / %d", s.WordlistPosition, s.WordlistSize)
	s.Mu.RUnlock()
}
