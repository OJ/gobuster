package gobuster

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
)

func PrepareSignalHandler(s *State) {
	s.SignalChan = make(chan os.Signal, 1)
	signal.Notify(s.SignalChan, os.Interrupt)
	go func() {
		for _ = range s.SignalChan {
			// caught CTRL+C
			if !s.Quiet {
				fmt.Println("[!] Keyboard interrupt detected, terminating.")
				s.Terminate = true
			}
		}
	}()
}

func Ruler(s *State) {
	if !s.Quiet {
		fmt.Println("=====================================================")
	}
}

func Banner(s *State) {
	if s.Quiet {
		return
	}

	fmt.Println("")
	fmt.Println("Gobuster v1.3                OJ Reeves (@TheColonial)")
	Ruler(s)
}

func ShowConfig(s *State) {
	if s.Quiet {
		return
	}

	if s != nil {
		fmt.Printf("[+] Mode         : %s\n", s.Mode)
		fmt.Printf("[+] Url/Domain   : %s\n", s.Url)
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

			if s.ProxyUrl != nil {
				fmt.Printf("[+] Proxy        : %s\n", s.ProxyUrl)
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
