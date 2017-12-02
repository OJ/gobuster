// Buster

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"unicode/utf8"
)

// A single result which comes from an individual web
// request.
type busterResult struct {
	entity string
	status int
	extra  string
	size   *int64
}

// Small helper to combine URL with URI then make a
// request to the generated location.
func get(cfg *config, url, uri, cookie string) (*int, *int64) {
	return getResponse(cfg, url+uri, cookie)
}

// Make a request to the given URL.
func getResponse(cfg *config, fullUrl, cookie string) (*int, *int64) {
	req, err := http.NewRequest("GET", fullUrl, nil)

	if err != nil {
		return nil, nil
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	if cfg.UserAgent != "" {
		req.Header.Set("User-Agent", cfg.UserAgent)
	}

	if cfg.Username != "" {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}

	resp, err := cfg.Client.Do(req)

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
		if err := resp.Body.Close(); err != nil {
			log.Printf("[!] problem closing the response body")
		}
	}()

	var length *int64 = nil

	if cfg.IncludeLength {
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

func prepareSignalHandler(cfg *config) {
	cfg.SignalChan = make(chan os.Signal, 1)
	signal.Notify(cfg.SignalChan, os.Interrupt)
	go func() {
		for range cfg.SignalChan {
			// caught CTRL+C
			if !cfg.Quiet {
				fmt.Println("[!] Keyboard interrupt detected, terminating.")
				cfg.Terminate = true
			}
		}
	}()
}

// runBuster the busting of the website with the given
// set of settings from the command line.
func runBuster(cfg *config) {

	printConfig(cfg)

	if cfg.Setup(cfg) == false {
		printRuler(cfg)
		return
	}

	prepareSignalHandler(cfg)

	// channels used for comms
	wordChan := make(chan string, cfg.Threads)
	resultChan := make(chan busterResult)

	// Use a wait group for waiting for all threads
	// to finish
	processorGroup := new(sync.WaitGroup)
	processorGroup.Add(cfg.Threads)
	printerGroup := new(sync.WaitGroup)
	printerGroup.Add(1)

	// Create goroutines for each of the number of threads
	// specified.
	for i := 0; i < cfg.Threads; i++ {
		go func() {
			for {
				word := <-wordChan

				// Did we reach the end? If so break.
				if word == "" {
					break
				}

				// Mode-specific processing
				cfg.Processor(cfg, word, resultChan)
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
			cfg.Printer(cfg, &r)
		}
		printerGroup.Done()
	}()

	var scanner *bufio.Scanner

	if cfg.StdIn {
		// Read directly from stdin
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		// Pull content from the wordlist
		wordlist, err := os.Open(cfg.Wordlist)
		if err != nil {
			panic("Failed to open wordlist")
		}
		defer wordlist.Close()

		// Lazy reading of the wordlist line by line
		scanner = bufio.NewScanner(wordlist)
	}

	var outputFile *os.File
	if cfg.OutputFileName != "" {
		outputFile, err := os.Create(cfg.OutputFileName)
		if err != nil {
			fmt.Printf("[!] Unable to write to %s, falling back to stdout.\n", cfg.OutputFileName)
			cfg.OutputFileName = ""
			cfg.OutputFile = nil
		} else {
			cfg.OutputFile = outputFile
		}
	}

	for scanner.Scan() {
		if cfg.Terminate {
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
	if cfg.OutputFile != nil {
		outputFile.Close()
	}
	printRuler(cfg)
}
