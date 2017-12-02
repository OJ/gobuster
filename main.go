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
//
// Please see THANKS file for contributors.
// Please see LICENSE file for license details.
//
//----------------------------------------------------

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/satori/go.uuid"
)

// A single result which comes from an individual web
// request.
type Result struct {
	Entity string
	Status int
	Extra  string
	Size   *int64
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

// Process the busting of the website with the given
// set of settings from the command line.
func Process(cfg *config) {

	printConfig(cfg)

	if cfg.Setup(cfg) == false {
		printRuler(cfg)
		return
	}

	PrepareSignalHandler(cfg)

	// channels used for comms
	wordChan := make(chan string, cfg.Threads)
	resultChan := make(chan Result)

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

func SetupDns(cfg *config) bool {
	// Resolve a subdomain that probably shouldn't exist
	guid := uuid.NewV4()
	wildcardIps, err := net.LookupHost(fmt.Sprintf("%s.%s", guid, cfg.Url))
	if err == nil {
		cfg.IsWildcard = true
		cfg.WildcardIps.addRange(wildcardIps)
		fmt.Println("[-] Wildcard DNS found. IP address(es): ", cfg.WildcardIps.string())
		if !cfg.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard DNS, specify the '-fw' switch.")
		}
		return cfg.WildcardForced
	}

	if !cfg.Quiet {
		// Provide a warning if the base domain doesn't resolve (in case of typo)
		_, err = net.LookupHost(cfg.Url)
		if err != nil {
			// Not an error, just a warning. Eg. `yp.to` doesn't resolve, but `cr.py.to` does!
			fmt.Println("[-] Unable to validate base domain:", cfg.Url)
		}
	}

	return true
}

func SetupDir(cfg *config) bool {
	guid := uuid.NewV4()
	wildcardResp, _ := get(cfg, cfg.Url, fmt.Sprintf("%s", guid), cfg.Cookies)

	if cfg.StatusCodes.contains(*wildcardResp) {
		cfg.IsWildcard = true
		fmt.Println("[-] Wildcard response found:", fmt.Sprintf("%s%s", cfg.Url, guid), "=>", *wildcardResp)
		if !cfg.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard responses, specify the '-fw' switch.")
		}
		return cfg.WildcardForced
	}

	return true
}

func ProcessDnsEntry(cfg *config, word string, resultChan chan<- Result) {
	subdomain := word + "." + cfg.Url
	ips, err := net.LookupHost(subdomain)

	if err == nil {
		if !cfg.IsWildcard || !cfg.WildcardIps.containsAny(ips) {
			result := Result{
				Entity: subdomain,
			}
			if cfg.ShowIPs {
				result.Extra = strings.Join(ips, ", ")
			} else if cfg.ShowCNAME {
				cname, err := net.LookupCNAME(subdomain)
				if err == nil {
					result.Extra = cname
				}
			}
			resultChan <- result
		}
	} else if cfg.Verbose {
		result := Result{
			Entity: subdomain,
			Status: 404,
		}
		resultChan <- result
	}
}

func ProcessDirEntry(cfg *config, word string, resultChan chan<- Result) {
	suffix := ""
	if cfg.UseSlash {
		suffix = "/"
	}

	// Try the DIR first
	dirResp, dirSize := get(cfg, cfg.Url, word+suffix, cfg.Cookies)
	if dirResp != nil {
		resultChan <- Result{
			Entity: word + suffix,
			Status: *dirResp,
			Size:   dirSize,
		}
	}

	// Follow up with files using each ext.
	for ext := range cfg.Extensions {
		file := word + cfg.Extensions[ext]
		fileResp, fileSize := get(cfg, cfg.Url, file, cfg.Cookies)

		if fileResp != nil {
			resultChan <- Result{
				Entity: file,
				Status: *fileResp,
				Size:   fileSize,
			}
		}
	}
}

func WriteToFile(cfg *config, output string) {
	_, err := cfg.OutputFile.WriteString(output)
	if err != nil {
		log.Panicf("[!] Unable to write to file %v", cfg.OutputFileName)
	}
}

func PrepareSignalHandler(cfg *config) {
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

func main() {
	state := ParseCmdLine()
	if state != nil {
		Process(state)
	}
}
