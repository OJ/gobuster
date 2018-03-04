package libgobuster

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

// PrepareSignalHandler ... Set up a SIGINT handler
func PrepareSignalHandler(s *State) {
	s.SignalChan = make(chan os.Signal, 1)
	signal.Notify(s.SignalChan, os.Interrupt)
	go func() {
		for range s.SignalChan {
			// caught CTRL+C
			if !s.Quiet {
				fmt.Println("[!] Keyboard interrupt detected, terminating.")
				s.Terminate = true
			}
		}
	}()
}

// Ruler ... Perform advanced screen I/O :>
func Ruler(s *State) {
	if !s.Quiet {
		fmt.Println("=====================================================")
	}
}

// Banner ... Print the Gobuster banner to the screen
func Banner(s *State) {
	if s.Quiet {
		return
	}

	fmt.Println("")
	fmt.Println("Gobuster v1.4.1              OJ Reeves (@TheColonial)")
	Ruler(s)
}

// ShowConfig ... Print the state to the screen
func ShowConfig(s *State) {
	if s.Quiet {
		return
	}

	if s != nil {
		fmt.Printf("[+] Mode         : %s\n", s.Mode)
		fmt.Printf("[+] URL/Domain   : %s\n", s.URL)
		fmt.Printf("[+] Threads      : %d\n", s.Threads)

		wordlist := "stdin (pipe)"
		if !s.StdIn {
			wordlist = s.Wordlist
		}
		fmt.Printf("[+] Wordlist     : %s\n", wordlist)

		if s.OutputFileName != "" {
			fmt.Printf("[+] Output file  : %s\n", s.OutputFileName)
		}

		if s.Mode == "dir" {
			fmt.Printf("[+] Status codes : %s\n", s.StatusCodes.Stringify())

			if s.ProxyURL != nil {
				fmt.Printf("[+] Proxy        : %s\n", s.ProxyURL)
			}

			if s.Cookies != "" {
				fmt.Printf("[+] Cookies      : %s\n", s.Cookies)
			}

			if s.UserAgent != "" {
				fmt.Printf("[+] User Agent   : %s\n", s.UserAgent)
			}

			if s.IncludeLength {
				fmt.Printf("[+] Show length  : true\n")
			}

			if s.Username != "" {
				fmt.Printf("[+] Auth User    : %s\n", s.Username)
			}

			if len(s.Extensions) > 0 {
				fmt.Printf("[+] Extensions   : %s\n", strings.Join(s.Extensions, ","))
			}

			if s.UseSlash {
				fmt.Printf("[+] Add Slash    : true\n")
			}

			if s.FollowRedirect {
				fmt.Printf("[+] Follow Redir : true\n")
			}

			if s.Expanded {
				fmt.Printf("[+] Expanded     : true\n")
			}

			if s.NoStatus {
				fmt.Printf("[+] No status    : true\n")
			}

			if s.Verbose {
				fmt.Printf("[+] Verbose      : true\n")
			}
		}

		Ruler(s)
	}
}

// Add ... Add an element to a set
func (set *StringSet) Add(s string) bool {
	_, found := set.Set[s]
	set.Set[s] = true
	return !found
}

// AddRange ... Add a list of elements to a set
func (set *StringSet) AddRange(ss []string) {
	for _, s := range ss {
		set.Set[s] = true
	}
}

// Contains ... Test if an element is in a set
func (set *StringSet) Contains(s string) bool {
	_, found := set.Set[s]
	return found
}

// ContainsAny ... Check if any of the elements exist
func (set *StringSet) ContainsAny(ss []string) bool {
	for _, s := range ss {
		if set.Set[s] {
			return true
		}
	}
	return false
}

// Stringify ... Stringify the set
func (set *StringSet) Stringify() string {
	values := []string{}
	for s := range set.Set {
		values = append(values, s)
	}
	return strings.Join(values, ",")
}

// Add ... Add an element to a set
func (set *IntSet) Add(i int) bool {
	_, found := set.Set[i]
	set.Set[i] = true
	return !found
}

// Contains ... Test if an element is in a set
func (set *IntSet) Contains(i int) bool {
	_, found := set.Set[i]
	return found
}

// Stringify ... Stringify the set
func (set *IntSet) Stringify() string {
	values := []string{}
	for s := range set.Set {
		values = append(values, strconv.Itoa(s))
	}
	return strings.Join(values, ",")
}
