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
	"strconv"
	"strings"
	"sync"
)

// Shim type for "set"
type IntSet struct {
	set map[int]bool
}

// Contains settings that are read in from the command
// line when the program is invoked.
type Settings struct {
	Threads     int
	Wordlist    string
	Url         string
	Cookies     string
	Extensions  []string
	StatusCodes IntSet
	ShowAll     bool
	UseSlash    bool
}

// A single result which comes from an individual web
// request.
type Result struct {
	Uri    string
	Status int
}

// Add an element to a set
func (set *IntSet) Add(i int) bool {
	_, found := set.set[i]
	set.set[i] = true
	return !found
}

// Test if an element is in a set
func (set *IntSet) Contains(i int) bool {
	_, found := set.set[i]
	return found
}

// Stringify the set
func (set *IntSet) Stringify() string {
	values := []string{}
	for s, _ := range set.set {
		values = append(values, strconv.Itoa(s))
	}
	return strings.Join(values, ",")
}

// Make a request to the given URL.
func MakeRequest(client *http.Client, url, cookie string) *int {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil
	}

	defer resp.Body.Close()

	return &resp.StatusCode
}

// Small helper to combine URL with URI then make a
// request to the generated location.
func GoGet(client *http.Client, url, uri, cookie string) *int {
	return MakeRequest(client, url+uri, cookie)
}

// Parse all the command line options into a settings
// instance for future use.
func ParseCmdLine() *Settings {
	var extensions string
	var codes string

	s := Settings{StatusCodes: IntSet{set: map[int]bool{}}}

	// Set up the variables we're interested in parsing.
	flag.IntVar(&s.Threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&s.Wordlist, "w", "", "Path to the wordlist")
	flag.StringVar(&codes, "s", "200,204,301,302,307", "Positive status codes")
	flag.StringVar(&s.Url, "u", "", "The target URL")
	flag.StringVar(&s.Cookies, "c", "", "Cookies to use for the requests")
	flag.StringVar(&extensions, "x", "", "File extension(s) to search for")
	flag.BoolVar(&s.ShowAll, "v", false, "Show all results (not just positives)")
	flag.BoolVar(&s.UseSlash, "f", false, "Append a forward-slash to each directory request")

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

	// status codes are comma seaprated
	if codes != "" {
		for _, c := range strings.Split(codes, ",") {
			i, err := strconv.Atoi(c)
			if err != nil {
				panic("Invalid status code given")
			}
			s.StatusCodes.Add(i)
		}
	}

	client := &http.Client{}
	if GoGet(client, s.Url, "", s.Cookies) == nil {
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

	client := &http.Client{}

	suffix := ""
	if s.UseSlash {
		suffix = "/"
	}

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

				// Try the DIR first
				dirResp := GoGet(client, s.Url, word+suffix, s.Cookies)
				if dirResp != nil {
					resultChan <- Result{
						Uri:    word + suffix,
						Status: *dirResp,
					}
				}

				// Follow up with files using each ext.
				for ext := range s.Extensions {
					file := word + s.Extensions[ext]
					fileResp := GoGet(client, s.Url, file, s.Cookies)
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
			if s.StatusCodes.Contains(r.Status) {
				// Only print results out if we find something
				// meaningful.
				fmt.Printf("Found: /%s (%d)\n", r.Uri, r.Status)
			} else if s.ShowAll {
				// Print out other results if the user wants to
				// see them.
				fmt.Printf("Result: /%s (%d)\n", r.Uri, r.Status)
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
	fmt.Println("Gobuster v0.2 (OJ Reeves @TheColonial)")
	fmt.Println("======================================")

	settings := ParseCmdLine()

	if settings == nil {
		return
	}

	fmt.Printf("[+] Url          : %s\n", settings.Url)
	fmt.Printf("[+] Threads      : %d\n", settings.Threads)
	fmt.Printf("[+] Wordlist     : %s\n", settings.Wordlist)
	fmt.Printf("[+] Status codes : %s\n", settings.StatusCodes.Stringify())

	if settings.Cookies != "" {
		fmt.Printf("[+] Cookies      : %s\n", settings.Cookies)
	}

	if len(settings.Extensions) > 0 {
		fmt.Printf("[+] Extensions   : %s\n", strings.Join(settings.Extensions, ","))
	}

	if settings.UseSlash {
		fmt.Printf("[+] Add Slash    : true\n")
	}

	if settings.ShowAll {
		fmt.Printf("[+] Dislpay all  : true\n")
	}

	fmt.Println("======================================")

	Process(settings)
}
