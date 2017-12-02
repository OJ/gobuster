// Reporter

package main

import (
	"fmt"
	"log"
	"strings"
)

func printDnsResult(cfg *config, br *busterResult) {
	output := ""
	if br.status == 404 {
		output = fmt.Sprintf("Missing: %s\n", br.entity)
	} else if cfg.ShowIPs {
		output = fmt.Sprintf("Found: %s [%s]\n", br.entity, br.extra)
	} else if cfg.ShowCNAME {
		output = fmt.Sprintf("Found: %s [%s]\n", br.entity, br.extra)
	} else {
		output = fmt.Sprintf("Found: %s\n", br.entity)
	}
	fmt.Printf("%s", output)

	if cfg.OutputFile != nil {
		WriteToFile(cfg, output)
	}
}

func printDirResult(cfg *config, br *busterResult) {
	output := ""

	// Prefix if we're in verbose mode
	if cfg.Verbose {
		if cfg.StatusCodes.contains(br.status) {
			output = "Found : "
		} else {
			output = "Missed: "
		}
	}

	if cfg.StatusCodes.contains(br.status) || cfg.Verbose {
		if cfg.Expanded {
			output += cfg.Url
		} else {
			output += "/"
		}
		output += br.entity

		if !cfg.NoStatus {
			output += fmt.Sprintf(" (Status: %d)", br.status)
		}

		if br.size != nil {
			output += fmt.Sprintf(" [Size: %d]", *br.size)
		}
		output += "\n"

		fmt.Printf(output)

		if cfg.OutputFile != nil {
			WriteToFile(cfg, output)
		}
	}
}

func printRuler(cfg *config) {
	if !cfg.Quiet {
		fmt.Println("=====================================================")
	}
}

func printBanner(cfg *config) {
	if cfg.Quiet {
		return
	}

	fmt.Println("")
	fmt.Println("Gobuster v1.3                OJ Reeves (@TheColonial)")
	printRuler(cfg)
}

func printConfig(cfg *config) {
	if cfg.Quiet {
		return
	}

	if cfg != nil {
		fmt.Printf("[+] Mode         : %s\n", cfg.Mode)
		fmt.Printf("[+] Url/Domain   : %s\n", cfg.Url)
		fmt.Printf("[+] Threads      : %d\n", cfg.Threads)

		wordlist := "stdin (pipe)"
		if !cfg.StdIn {
			wordlist = cfg.Wordlist
		}
		fmt.Printf("[+] Wordlist     : %s\n", wordlist)

		if cfg.OutputFileName != "" {
			fmt.Printf("[+] Output file  : %s\n", cfg.OutputFileName)
		}

		if cfg.Mode == "dir" {
			fmt.Printf("[+] Status codes : %s\n", cfg.StatusCodes.string())

			if cfg.ProxyUrl != nil {
				fmt.Printf("[+] Proxy        : %s\n", cfg.ProxyUrl)
			}

			if cfg.Cookies != "" {
				fmt.Printf("[+] Cookies      : %s\n", cfg.Cookies)
			}

			if cfg.UserAgent != "" {
				fmt.Printf("[+] User Agent   : %s\n", cfg.UserAgent)
			}

			if cfg.IncludeLength {
				fmt.Printf("[+] Show length  : true\n")
			}

			if cfg.Username != "" {
				fmt.Printf("[+] Auth User    : %s\n", cfg.Username)
			}

			if len(cfg.Extensions) > 0 {
				fmt.Printf("[+] Extensions   : %s\n", strings.Join(cfg.Extensions, ","))
			}

			if cfg.UseSlash {
				fmt.Printf("[+] Add Slash    : true\n")
			}

			if cfg.FollowRedirect {
				fmt.Printf("[+] Follow Redir : true\n")
			}

			if cfg.Expanded {
				fmt.Printf("[+] Expanded     : true\n")
			}

			if cfg.NoStatus {
				fmt.Printf("[+] No status    : true\n")
			}

			if cfg.Verbose {
				fmt.Printf("[+] Verbose      : true\n")
			}
		}

		printRuler(cfg)
	}
}

func WriteToFile(cfg *config, output string) {
	_, err := cfg.OutputFile.WriteString(output)
	if err != nil {
		log.Panicf("[!] Unable to write to file %v", cfg.OutputFileName)
	}
}
