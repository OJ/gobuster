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
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

// Contains settings that are read in from the command
// line when the program is invoked.
type Settings struct {
	Threads    int
	Wordlist   string
	Url        string
	Extensions []string
	ShowAll    bool
}

// A single result which comes from an individual web
// request.
type Result struct {
	Uri    string
	Status int
}

// Make a request to the given URL.
func MakeRequest(url string) *int {
	resp, err := http.Get(url)

	if err != nil {
		return nil
	}

	defer resp.Body.Close()

	return &resp.StatusCode
}

// Small helper to combine URL with URI then make a
// request to the generated location.
func GoGet(url, uri string) *int {
	return MakeRequest(url + uri)
}

// Parse all the command line options into a settings
// instance for future use.
func ParseCmdLine() *Settings {
	var extensions string
	s := Settings{}

	// Set up the variables we're interested in parsing.
	flag.IntVar(&s.Threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&s.Wordlist, "w", "", "Path to the wordlist")
	flag.StringVar(&s.Url, "u", "", "The target URL")
	flag.StringVar(&extensions, "x", "", "File extension(s) to search for")
	flag.BoolVar(&s.ShowAll, "s", false, "Show all results (not just positives)")

	flag.Parse()

	if s.Threads < 0 {
		fmt.Println("Threads (-t): Invalid value", s.Threads)
		return nil
	}

	if s.Wordlist == "" {
		fmt.Println("WordList (-w): Must be specified")
		return nil
	}

	if _, err := os.Stat(s.Wordlist); os.IsNotExist(err) {
		fmt.Println("Wordlist (-w): File does not exist", s.Wordlist)
		return nil
	}

	if s.Url == "" {
		fmt.Println("Url (-u): Must be specified")
		return nil
	}

	if strings.HasSuffix(s.Url, "/") == false {
		s.Url = s.Url + "/"
	}

	// extensions are comma seaprated
	if extensions != "" {
		s.Extensions = strings.Split(extensions, ",")
	}

	if GoGet(s.Url, "") == nil {
		fmt.Println("Url (-u): Unable to connect", s.Url)
	}

	return &s
}

// Process the busting of the website with the given
// set of settings from the command line.
func Process(s *Settings) {
	wordlist, err := os.Open(s.Wordlist)
	if err != nil {
		panic("Failed to open wordlist")
	}

	// channels used for comms
	wordChan := make(chan string, s.Threads)
	resultChan := make(chan Result)

	// Use a wait group for waiting for all threads
	// to finish
	wg := new(sync.WaitGroup)
	wg.Add(s.Threads)

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

				// Try the DIR first.
				dirResp := GoGet(s.Url, word)
				if dirResp != nil {
					resultChan <- Result{
						Uri:    word,
						Status: *dirResp,
					}
				}

				// Follow up with files using each ext.
				for ext := range s.Extensions {
					file := word + s.Extensions[ext]
					fileResp := GoGet(s.Url, file)
					if fileResp != nil {
						resultChan <- Result{
							Uri:    file,
							Status: *fileResp,
						}
					}
				}
			}

			// Indicate to the wait group that the thread
			// has finished.
			wg.Done()
		}()
	}

	// Single goroutine which handles the results as they
	// appear from the worker threads.
	go func() {
		for r := range resultChan {
			switch r.Status {
			// Only print results out if we find something
			// meaningful.
			case 200, 204, 301, 302, 307:
				fmt.Printf("Found: /%s (%d)\n", r.Uri, r.Status)
			default:
				// Print out other results if the user wants to
				// see them.
				if s.ShowAll {
					fmt.Printf("Result: /%s (%d)\n", r.Uri, r.Status)
				}
			}
		}
	}()

	defer wordlist.Close()

	// Lazy reading of the wordlist line by line
	scanner := bufio.NewScanner(wordlist)
	for scanner.Scan() {
		word := scanner.Text()

		// Skipe "comment" lines
		if strings.HasPrefix(word, "#") == false {
			wordChan <- word
		}
	}

	close(wordChan)
	wg.Wait()
	close(resultChan)
}

func main() {
	fmt.Println("Gobuster v0.1 (OJ Reeves @TheColonial)")
	fmt.Println("======================================")

	settings := ParseCmdLine()

	if settings == nil {
		return
	}

	fmt.Printf("[+] Url        : %s\n", settings.Url)
	fmt.Printf("[+] Threads    : %d\n", settings.Threads)
	fmt.Printf("[+] Wordlist   : %s\n", settings.Wordlist)

	if len(settings.Extensions) > 0 {
		fmt.Printf("[+] Extensions : %s\n", strings.Join(settings.Extensions, ","))
	}

	fmt.Println("======================================")

	Process(settings)
}
