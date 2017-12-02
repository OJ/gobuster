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
	} else if cfg.showIPs {
		output = fmt.Sprintf("Found: %s [%s]\n", br.entity, br.extra)
	} else if cfg.showCNAME {
		output = fmt.Sprintf("Found: %s [%s]\n", br.entity, br.extra)
	} else {
		output = fmt.Sprintf("Found: %s\n", br.entity)
	}
	fmt.Printf("%s", output)

	if cfg.outputFile != nil {
		WriteToFile(cfg, output)
	}
}

func printDirResult(cfg *config, br *busterResult) {
	output := ""

	// Prefix if we're in verbose mode
	if cfg.verbose {
		if cfg.statusCodes.contains(br.status) {
			output = "Found : "
		} else {
			output = "Missed: "
		}
	}

	if cfg.statusCodes.contains(br.status) || cfg.verbose {
		if cfg.expanded {
			output += cfg.url
		} else {
			output += "/"
		}
		output += br.entity

		if !cfg.noStatus {
			output += fmt.Sprintf(" (Status: %d)", br.status)
		}

		if br.size != nil {
			output += fmt.Sprintf(" [Size: %d]", *br.size)
		}
		output += "\n"

		fmt.Printf(output)

		if cfg.outputFile != nil {
			WriteToFile(cfg, output)
		}
	}
}

func printRuler(cfg *config) {
	if !cfg.quiet {
		fmt.Println("=====================================================")
	}
}

func printBanner(cfg *config) {
	if cfg.quiet {
		return
	}

	fmt.Println("")
	fmt.Println("Gobuster v1.3                OJ Reeves (@TheColonial)")
	printRuler(cfg)
}

func printConfig(cfg *config) {
	if cfg.quiet {
		return
	}

	if cfg != nil {
		fmt.Printf("[+] Mode         : %s\n", cfg.mode)
		fmt.Printf("[+] Url/Domain   : %s\n", cfg.url)
		fmt.Printf("[+] Threads      : %d\n", cfg.threads)

		wordlist := "stdin (pipe)"
		if !cfg.stdIn {
			wordlist = cfg.wordlist
		}
		fmt.Printf("[+] Wordlist     : %s\n", wordlist)

		if cfg.outputFileName != "" {
			fmt.Printf("[+] Output file  : %s\n", cfg.outputFileName)
		}

		if cfg.mode == "dir" {
			fmt.Printf("[+] Status codes : %s\n", cfg.statusCodes.string())

			if cfg.proxyUrl != nil {
				fmt.Printf("[+] Proxy        : %s\n", cfg.proxyUrl)
			}

			if cfg.cookies != "" {
				fmt.Printf("[+] Cookies      : %s\n", cfg.cookies)
			}

			if cfg.userAgent != "" {
				fmt.Printf("[+] User Agent   : %s\n", cfg.userAgent)
			}

			if cfg.includeLength {
				fmt.Printf("[+] Show length  : true\n")
			}

			if cfg.username != "" {
				fmt.Printf("[+] Auth User    : %s\n", cfg.username)
			}

			if len(cfg.extensions) > 0 {
				fmt.Printf("[+] Extensions   : %s\n", strings.Join(cfg.extensions, ","))
			}

			if cfg.useSlash {
				fmt.Printf("[+] Add Slash    : true\n")
			}

			if cfg.followRedirect {
				fmt.Printf("[+] Follow Redir : true\n")
			}

			if cfg.expanded {
				fmt.Printf("[+] Expanded     : true\n")
			}

			if cfg.noStatus {
				fmt.Printf("[+] No status    : true\n")
			}

			if cfg.verbose {
				fmt.Printf("[+] Verbose      : true\n")
			}
		}

		printRuler(cfg)
	}
}

func WriteToFile(cfg *config, output string) {
	_, err := cfg.outputFile.WriteString(output)
	if err != nil {
		log.Panicf("[!] Unable to write to file %v", cfg.outputFileName)
	}
}
